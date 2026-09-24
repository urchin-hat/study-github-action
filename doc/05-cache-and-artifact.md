# Chapter 05: データの受け渡し（CacheとArtifact）

対応Issue: [#8 05: CacheとArtifactを使い分ける](https://github.com/urchin-hat/study-github-action/issues/8)

## この章で理解したいこと

- **Artifact（成果物）**:
  - `actions/upload-artifact` によるビルド成果物の保存
  - Web UIからのダウンロード確認
  - `actions/download-artifact` による別Jobへのファイル受け渡しと検証
  - 保持期間（`retention-days`）とストレージ容量・課金への影響
- **Cache（キャッシュ）**:
  - `actions/cache` による依存関係やビルドキャッシュの保存とリストア
  - Cache Hit と Cache Miss の挙動
  - Cache Miss を「正常系」として扱う耐障害設計
- **GitLab CI/CDとの違いと、Cache vs Artifact の使い分け基準**

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 観点 |
| --- | --- | --- |
| `artifacts: paths: [...]` | `actions/upload-artifact@v4` | 成果物の保存・Webからのダウンロード |
| `dependencies:` または自動引き継ぎ | `actions/download-artifact@v4` | 後続Jobでの成果物ダウンロード |
| `artifacts: expire_in: 1 week` | `with: retention-days: 1` | 成果物の保持期間 |
| `cache: paths: [...]` / `cache: key:` | `actions/cache@v4` (`path`, `key`, `restore-keys`) | 高速化用の一時キャッシュ |
| キャッシュが消えても動く設計 | キャッシュが消えても動く設計（Cache Miss正常系） | キャッシュの基本原則 |

GitLab CI/CDでは `artifacts:` と書くだけで後続ステージのJobに自動でファイルがダウンロードされますが、GitHub Actionsでは **アップロード（`upload-artifact`）もダウンロード（`download-artifact`）もActionで明示的に呼び出す** 必要があります。

## 最初の方針

1. `build` Jobで生成したバイナリ（`bin/study-server`）を `actions/upload-artifact` でアップロードする
2. 後続の新しいJob（例: `verify` や `e2e-test`）を作成し、`actions/download-artifact` でバイナリを取得して実行・検証する
3. `retention-days`（保持期間）を設定してみる
4. `actions/cache` を使ってキャッシュの生成・復元（Cache Hit / Cache Miss）を確認する

## 実装前の予想

- [x] GitLab CI/CDでは先行JobのArtifactは後続Jobに自動で展開されるが、GitHub Actionsではどうなるか？: 予想は自動展開（A）。実際は明示的なアクション（`actions/download-artifact@v4`）の呼び出しが必要（B）。勝手に帯域やディスクを消費しないGitHub Actionsの明示的設計によるもの。
- [x] CacheとArtifactの最大の違いは何か？（消えたらどうなるか？）:
  - 予想: ArtifactがX（消えても平気）、CacheがY（消えたら困る）。
  - 実際: **実は真逆！**
    - **Cache = X（消えてもOK・再生成可能）**: 依存ライブラリなど。容量逼迫時にGitHubが勝手に削除（eviction）するため、消えても再取得してCIが完走する（Cache Missが正常系）設計が必須。
    - **Artifact = Y（消えたら困る・確定成果物）**: このコミットでビルドしたバイナリなど。後続デプロイやリリースに必要なため、消えると後続が動けない確定成果物。
- [ ] Artifactのデフォルトの保存期間（保持日数）は何日か？

## 壁打ちメモ

### 2026-09-25: Chapter 05開始
- `main` からブランチ `lesson/05-cache-and-artifact` を作成。

### 2026-09-25: ArtifactとCacheのメンタルモデル（予想と壁打ち）

1. **後続Jobへの引き継ぎ**:
   - 予想: GitLab CI/CDのように自動でダウンロード・展開される（A）。
   - 実際: **B（明示的に `actions/download-artifact` が必要）**。
   - `actions/checkout` と同様、GitHub Actionsでは「必要なものだけを明示的に呼び出す」というオプトイン思想が徹底されている。
2. **CacheとArtifactの本質的な違い（消えたときの挙動）**:
   - 予想: ArtifactがX（消えても平気）、CacheがY（消えたら困る）。
   - 実際: **真逆** である。
     - **Cache（高速化のための一時データ）**:
       - あくまで「前回の結果を使い回して実行時間を短縮する」ためのもの。
       - GitHubの容量制限（リポジトリあたり10GB）や保持期限（7日間アクセスなし）でいつでもGitHub側に勝手にパージ（削除）される運命にある。
       - そのため、**「Cache Miss（キャッシュが存在しないこと）は日常茶飯事の正常系」** として設計しなければならない。
     - **Artifact（パイプラインの確定成果物）**:
       - ビルドバイナリ、テストレポート、リリース用パッケージなど。
       - 「そのコミットでビルドされた正真正銘の成果物」であり、消えてしまったら後続のデプロイや検証は成立しない。
       - 明示的にアップロード・ダウンロードされ、確実に保持される。

## 試したことと結果

### 実験1: `upload-artifact` と `download-artifact` によるバイナリの保存とJob間受け渡し

- PR: [#18](https://github.com/urchin-hat/study-github-action/pull/18)
- Workflow Run: [36021645589](https://github.com/urchin-hat/study-github-action/actions/runs/36021645589)
- 目的:
  - `build` Jobで生成したバイナリ `bin/study-server` を `actions/upload-artifact@v4` でアップロードする。
  - 後続の `verify` Jobで `actions/download-artifact@v4` を使ってバイナリを取得する。
  - `verify` JobにはGo環境やソースコードをチェックアウトせず、ダウンロードしたバイナリ単体を起動してHTTPリクエスト（`/health`, `/hello`）が正常に通るかを検証する。
  - `retention-days: 1` で保持期間が設定されることを確認する。

#### 実行結果
- 各Jobのステータス:
  - `Build Binary`: ✅ **success** (22秒) — バイナリをビルドし、`upload-artifact` でアップロード完了（ZIP圧縮後 4.2MB）。
  - `Verify Artifact`: ✅ **success** (7秒) — `download-artifact` で取得し、バイナリを起動してヘルスチェック成功！
- Web UI / CLIの確認:
  - Runの成果物一覧に `study-server-binary` が登録され、ZIPとしてダウンロード可能になった。
- `Verify Artifact` Jobのログ出力:
  ```text
  === Check downloaded files ===
  -rw-r--r-- 1 runner runner 7306544 Sep 24 15:39 study-server

  === Add execute permission and run binary ===
  2026/09/24 15:39:39 Starting server on :8080

  === Test /health endpoint ===
  HTTP/1.1 200 OK
  Content-Type: application/json
  {"status":"ok"}

  === Test /hello endpoint ===
  HTTP/1.1 200 OK
  Hello, GitHubActions!
  Verification passed successfully!
  ```

#### 分かったこと
- **Job間の確実な成果物受け渡し**:
  - `verify` Jobには `actions/checkout` も `actions/setup-go` も書かず、完全な素のUbuntu環境でバイナリ単体をダウンロードして動かすことができた。デプロイジョブと同じ挙動を完璧に再現できた。
- **実行権限（パーミッション）の落とし穴**:
  - `download-artifact` でダウンロードしたバイナリのパーミッションは `-rw-r--r--`（実行権限なし）になっていた。
  - アーティファクトは内部的にZIPで転送・保存されるため、実行権限（`+x`）が落ちる場合がある。実務でバイナリを扱う際は `chmod +x` を呼ぶ必要がある。
- **保持期間（retention-days）によるストレージ保護**:
  - `retention-days: 1` を明示することで、デフォルトの90日間保持による不要なストレージ消費・課金を防ぐことができる。

## つまずいた点

（実験を通して記録します）

## ブログへ残したい要点

（実験を通して記録します）

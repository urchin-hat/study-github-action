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

- [ ] GitLab CI/CDでは先行JobのArtifactは後続Jobに自動で展開されるが、GitHub Actionsではどうなるか？
- [ ] CacheとArtifactの最大の違いは何か？（消えたらどうなるか？）
- [ ] Artifactのデフォルトの保存期間（保持日数）は何日か？

## 壁打ちメモ

### 2026-09-25: Chapter 05開始
- `main` からブランチ `lesson/05-cache-and-artifact` を作成。

## 試したことと結果

（実験を段階的に実施して記録していきます）

## つまずいた点

（実験を通して記録します）

## ブログへ残したい要点

（実験を通して記録します）

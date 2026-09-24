# Chapter 04: サンプルアプリのCIを作る（Go）

対応Issue: [#7 04: サンプルアプリのCIを作る](https://github.com/urchin-hat/study-github-action/issues/7)

## この章で理解したいこと

- 公式Action（`actions/checkout`, `actions/setup-go`）を使った実コードの取得と言語ランタイムのセットアップ
- GitLab CI/CDのコンテナイメージ指定（`image:`）とGitHub Actionsの「ホストRunner + Setup Action」のアプローチの違い
- `lint`, `test`, `build` の実務的なCI Job分離とDAG依存関係（`needs`）
- `strategy.matrix` による複数言語バージョンの並列テスト
- `fail-fast` の挙動（デフォルトのキャンセル動作と制御）

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 観点 |
| --- | --- | --- |
| `image: golang:1.22` | `runs-on: ubuntu-latest` + `actions/setup-go@v5` | 実行環境・言語ランタイムの準備 |
| リポジトリは自動Clone | `actions/checkout@v4` が明示的に必要 | ソースコードの取得 |
| `parallel: matrix` | `strategy: matrix` | 複数バージョン等の組み合わせ並列実行 |
| `parallel: matrix` は個別完走 | `strategy.fail-fast: true`（既定値） | matrixの1つがコケた時の他ジョブのキャンセル |
| `cache: paths: [/go/pkg/mod]` | `actions/setup-go` 内蔵キャッシュ | 言語依存関係のキャッシュ |

GitLab CI/CDではJobごとにDockerコンテナイメージ（`image: golang:1.22`）を直接指定してその中でコマンドを実行するのが標準的ですが、GitHub Actionsでは **VM型Runner（`ubuntu-latest`）の上で公式の `setup-*` Actionを使って言語環境を構成する** のが標準的なアプローチです。

## 最初の方針

1. サンプルアプリ（Go言語の最小Webサーバー＋ユニットテスト）を作成する
2. `actions/checkout` と `actions/setup-go` を使って、リポジトリのコード取得とGo実行環境のセットアップを行う
3. `lint`（または `go vet` / `test`）、`build` のJobを分離して組み立てる
4. `strategy.matrix` を導入し、Goの複数バージョン（例: 1.22, 1.23）で並列テストを実行する
5. `fail-fast` の挙動（デフォルトで他がどうなるか）を検証する

## 実装前の予想

- [ ] リポジトリのコードは、GitLab CI/CDのようにJob開始時に自動でRunner上にCloneされているか？それとも明示的な手順が必要か？
- [ ] `actions/setup-go` を使う際、Goのバージョン指定はどう書くか？
- [ ] `strategy.matrix` で複数バージョンを指定した場合、各バージョンは直列で動くか？並列で動くか？
- [ ] Matrixの1つが失敗した場合、実行中の他のMatrixジョブはどうなるか？（`fail-fast` の既定値）

## 壁打ちメモ

### 2026-09-24: Chapter 04開始
- `main` からブランチ `lesson/04-application-ci` を作成。

## 試したことと結果

（実験を段階的に実施して記録していきます）

## つまずいた点

（実験を通して記録します）

## ブログへ残したい要点

（実験を通して記録します）

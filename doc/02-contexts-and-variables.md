# Chapter 02: Contextと変数

対応Issue: [#5 Contextと変数を学ぶ](https://github.com/urchin-hat/study-github-action/issues/5)

## この章で理解したいこと

- GitHub Actionsのexpression (`${{ ... }}`) とshell環境変数 (`$VAR`) の違い
- 主要なContext (`github`, `env` など) の役割と中身
- 環境変数のscope（Workflow全体、Job単位、Step単位）と優先順位
- `if` 条件式でのContextの利用
- 未信頼な入力（Context）をshell scriptに直接展開する危険性と安全な渡し方

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 観点 |
| --- | --- | --- |
| Predefined CI/CD variables (`CI_COMMIT_REF_NAME` 等) | GitHub Context (`github.ref_name` 等) | 実行環境やGitの状態を表す事前定義された値 |
| `variables` (pipeline/job scope) | `env` (workflow/job/step scope) | 定義した環境変数の有効範囲 |
| `rules:if` での変数評価 | Job/Stepの `if` 条件式 | 処理を実行するかどうかの条件分岐 |
| Script内での `$VAR` 参照 | Step `run` 内での `$VAR` 参照 | Shellレベルでの環境変数参照 |

GitLab CI/CDでは、定義した `variables` も事前定義変数も基本的にはすべてJob実行時に「環境変数」として注入される。
一方、GitHub Actionsでは **GitHubが解釈・置換する「Context/Expression」** と **RunnerのOS/Shellが扱う「環境変数」** が明確に分かれている。

## 最初の方針

1. Expressionによる直接展開（`${{ ... }}`）と環境変数（`$VAR`）の動作・ログ出力を比較する
2. Workflow、Job、Stepの各階層で `env` を定義し、スコープと優先順位（シャドーイング）を確認する
3. Stepの `if` 条件式を使い、条件分岐の挙動を確認する
4. 未信頼入力の直接埋め込みリスク（スクリプトインジェクション）を実験・確認する

## 実装前の予想

- [x] 表記の違い: ContextはJinjaに似た `${{ ... }}` テンプレート構文、シェル環境変数は `$VAR`
- [x] 評価タイミング: Expression (`${{ ... }}`) はシェル起動前のワークフロー解釈時に展開される
- [ ] Workflow全体、Job、Stepで同名の環境変数を定義した場合、優先順位はどうなるか？
- [ ] `if` 条件式の中では `${{ ... }}` を囲む必要があるか？省略できるか？
- [ ] Contextの値を `run` 内のスクリプトで直接 `${{ ... }}` と書く場合と、一度 `env` に渡して `$VAR` で参照する場合の安全性の違いは何か？

## 壁打ちメモ

### 2026-09-24: Chapter 02開始

Section 01のPR #14が`main`へmergeされたことを受け、`main`から`lesson/02-contexts-and-variables`を作成した。

### 2026-09-24: Expressionと環境変数の違いを予想

1. 表記の違い:
   - 予想: Jinjaのようなテンプレート構文を使う
   - 実際: `${{ <expression> }}` 構文を使う。シェル変数 `$VAR` とは明確に異なる。
2. 評価タイミング:
   - 予想: 起動タイミング（シェル実行前）
   - 実際: そのとおり。Runnerがシェルプロセスを起動する前にGitHub Actionsエンジンが `${{ ... }}` を文字列置換する。
3. 実行ログでの表示予想:
   - Expression展開: Runnerがシェル起動前に置換するため、ログの実行コマンド表示は値が展開された `echo "Value: Hello from env"` になる
   - シェル環境変数: シェルプロセスが環境変数を展開するため、ログの実行コマンド表示は変数のまま `echo "Value: $MESSAGE"` になる


## 試したことと結果

壁打ちと実装の進行に合わせて追記する。

## つまずいた点

- **`${{ ... }}` を文字列として書くと構文エラーになる**:
  - `echo "=== 1. Expression展開 (${{ ... }}) ==="` と記述したところ、GitHub Actionsのパーサーが `${{ ... }}` をExpression式として評価しようとし、`(Line: 22, Col: 14): Unexpected symbol: '...'. Located at position 1 within expression: ...` で即座にエラーとなった。
  - ワークフローファイル内でリテラルとして `${{` を出力したい場合は `${{ '${{ ... }}' }}` のようにエスケープするか、記号を避ける必要がある。
- **構文エラーがある場合のWorkflow Runの挙動**:
  - `push.branches: [main]` で作業ブランチへのpushが除外されている設定であっても、ワークフロー定義に構文エラーがあると、ブランチ判定以前にGitHubが即座に failure（Workflow file issue）のRunを作成する。


## ブログへ残したい要点

壁打ちと実装の進行に合わせて追記する。

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
- [x] Workflow全体、Job、Stepで同名の環境変数を定義した場合、優先順位はどうなるか？: Step -> Job -> Workflow（狭いスコープが優先）
- [x] `if` 条件式の中では `${{ ... }}` を囲む必要があるか？省略できるか？: 省略可能（公式推奨）。文字列リテラルはシングルクォートのみ有効。
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
   - コマンド表示: Expression展開は値置換済み（`echo "Value: Hello from env"`）、シェル環境変数は変数名のまま（`echo "Value: $MESSAGE"`）。
   - 標準出力: どちらも `Value: Hello from env` と出力された。シェル（bash）がダブルクォート内の環境変数を展開するため、出力結果は同じになる。結果が同じでも「誰が・いつ展開しているか」が本質的な違い。

### 2026-09-24: 環境変数のスコープと優先順位を予想

1. 優先順位（シャドーイング）:
   - 予想: `Step -> Job -> Workflow` の順で評価される（より狭いスコープが優先される）。
2. スコープ（有効範囲）:
   - 予想: あるStepで定義した `env` は、後続のStepからは参照されない。

### 2026-09-24: `if` 条件式の記法とクォートを予想

1. `${{ ... }}` の省略可否:
   - 予想: 省略できる（bashと同じなら）。
   - 実際: そのとおり。`if:` に書かれた式は自動的にExpressionとして評価されるため、`${{ ... }}` を省略して書くのが推奨。
2. 文字列のクォート:
   - 予想: どちらでも良いがダブルクォートが良い。
   - 実際: GitHub ActionsのExpression言語では **シングルクォート `'...'` のみ** が文字列リテラルとして定義されている。ダブルクォート `"..."` を使うとYAMLのパースと干渉したり、未定義の識別子としてエラーになる落とし穴がある。




## 試したことと結果

### 実験1: Expression展開とシェル環境変数の比較

- PR: [#15](https://github.com/urchin-hat/study-github-action/pull/15)
- Workflow Run: [35972524271](https://github.com/urchin-hat/study-github-action/actions/runs/35972524271)

実行ログのコマンド展開部分：
```text
Run echo "=== 1. Expression展開 ==="
echo "=== 1. Expression展開 ==="
echo "Value: Hello from env"
echo "Event: pull_request"
echo "=== 2. シェル環境変数 ==="
echo "Value: $MESSAGE"
echo "Event: $GITHUB_EVENT_NAME"
```

実行出力結果：
```text
=== 1. Expression展開 ===
Value: Hello from env
Event: pull_request
=== 2. シェル環境変数 ===
Value: Hello from env
Event: pull_request
```

#### 分かったこと
- 予想どおり、**Expression（`${{ ... }}`）はRunnerがシェルを起動する前に値へ文字列置換される**ため、シェルに渡されるコマンド自体が `echo "Value: Hello from env"` に書き換わっていた。
- 一方、**シェル環境変数（`$MESSAGE` や `$GITHUB_EVENT_NAME`）はシェルに `$VAR` のまま渡り、実行時にシェル自身が環境変数テーブルを参照して展開**していた。
- 最終的な出力文字列は同じだが、シェルに渡る「コードそのもの」が異なっていることが確認できた。

### 実験2: 環境変数のスコープと優先順位（シャドーイング）

- PR: [#15](https://github.com/urchin-hat/study-github-action/pull/15)
- Workflow Run: [35973010679](https://github.com/urchin-hat/study-github-action/actions/runs/35973010679)

実行出力結果：
```text
=== Step 1 ===
LEVEL: step
SHARED: from-workflow
JOB_ONLY: from-job
STEP_ONLY: from-step1

=== Step 2 ===
LEVEL: job
STEP_ONLY: ''
```

#### 分かったこと
- **優先順位（シャドーイング）**: `Step -> Job -> Workflow` の順で、より狭いスコープで定義された環境変数が優先して上書きされる（Step 1 では `LEVEL=step` が勝つ）。
- **スコープ（有効範囲）**:
  - 外側のスコープ（WorkflowやJob）で定義された変数は、内側のStepへ引き継がれる（`SHARED`、`JOB_ONLY` がStep 1 で参照可能）。
  - Stepレベルで定義された `env`（`STEP_ONLY` や Step 1 の `LEVEL=step`）は、**そのStepのプロセス内でのみ有効**であり、後続のStep 2 では `STEP_ONLY` は空になり、`LEVEL` もJobレベルの `job` に戻る。



## つまずいた点

- **`${{ ... }}` を文字列として書くと構文エラーになる**:
  - `echo "=== 1. Expression展開 (${{ ... }}) ==="` と記述したところ、GitHub Actionsのパーサーが `${{ ... }}` をExpression式として評価しようとし、`(Line: 22, Col: 14): Unexpected symbol: '...'. Located at position 1 within expression: ...` で即座にエラーとなった。
  - ワークフローファイル内でリテラルとして `${{` を出力したい場合は `${{ '${{ ... }}' }}` のようにエスケープするか、記号を避ける必要がある。
- **構文エラーがある場合のWorkflow Runの挙動**:
  - `push.branches: [main]` で作業ブランチへのpushが除外されている設定であっても、ワークフロー定義に構文エラーがあると、ブランチ判定以前にGitHubが即座に failure（Workflow file issue）のRunを作成する。


## ブログへ残したい要点

壁打ちと実装の進行に合わせて追記する。

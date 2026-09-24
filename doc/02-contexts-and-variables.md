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
- [x] Contextの値を `run` 内のスクリプトで直接 `${{ ... }}` と書く場合と、一度 `env` に渡して `$VAR` で参照する場合の安全性の違いは何か？: 直接展開するとシェルコードとして解釈されスクリプトインジェクションが発生する。`env` 経由にすると「コード」ではなく「データ（環境変数）」として渡るため無害化される。

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

### 2026-09-24: スクリプトインジェクションの危険性と環境変数経由の安全性を予想

1. 直接埋め込みの危険性:
   - 予想: インジェクションが起きる。
   - 実際: シェル起動前に文字列置換されるため、悪意ある引用符やセミコロン・`&&` が「シェルの構文」としてそのまま解釈され、任意のコマンドが実行されてしまう。
2. 環境変数経由の安全性:
   - 予想: 変数として評価されるため安全になる。
   - 実際: `env:` に渡すと、値はシェルスクリプトの「コード」ではなく「プロセスの環境変数（データ）」としてメモリ上に配置される。ダブルクォートで `"$VAR"` と参照すれば、シェルは文字列引数（データ）として扱い、内部の記号をコマンドとして実行しない。





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

### 実験3: `if` 条件式によるStep実行制御とExpression記法

- PR: [#15](https://github.com/urchin-hat/study-github-action/pull/15)
- Workflow Run: [35973428074](https://github.com/urchin-hat/study-github-action/actions/runs/35973428074)

実行結果：
- **Step A (`if: github.event_name == 'pull_request'` / 省略記法)**:
  - 実行された（`This step runs because event is pull_request`）
- **Step B (`if: github.event_name == 'push'` / 条件不一致)**:
  - **スキップされた**（ログ自体が生成されず、ステップが実行されなかった）
- **Step C (`if: ${{ github.event_name == 'pull_request' }}` / 明示的Expression記法)**:
  - 実行された（`This step also runs with explicit expression syntax`）

#### 分かったこと
- `if:` 条件式では、`${{ ... }}` を省略しても明示しても同様に評価されるが、省略記法がシンプルで推奨される。
- 条件式が `false` と評価されたStepは安全にスキップされる。

### 実験4: スクリプトインジェクションの実証と安全な環境変数経由の取り扱い

- PR: [#15](https://github.com/urchin-hat/study-github-action/pull/15)
- Workflow Run: [35973825486](https://github.com/urchin-hat/study-github-action/actions/runs/35973825486)
- 検証用PRタイトル: `Chapter 02: Context"; echo "=== INJECTION OCCURRED ==="; echo "`

#### 1. 直接展開（脆弱な例）の実行ログ
```text
Run echo "PR Title is: Chapter 02: Context"; echo "=== INJECTION OCCURRED ==="; echo ""
PR Title is: Chapter 02: Context
=== INJECTION OCCURRED ===
```
- **結果**: シェル起動前の文字列置換により、PRタイトルに含まれるダブルクォートとセミコロンがシェルのコマンド区切り文字として認識され、`echo "=== INJECTION OCCURRED ==="` が**意図しない独立したコマンドとして実行（インジェクション成功）**してしまった。

#### 2. 環境変数経由（安全な例）の実行ログ
```text
Run echo "PR Title is: $PR_TITLE"
env:
  PR_TITLE: Chapter 02: Context"; echo "=== INJECTION OCCURRED ==="; echo "
出力:
PR Title is: Chapter 02: Context"; echo "=== INJECTION OCCURRED ==="; echo "
```
- **結果**: シェルに渡るスクリプトは `echo "PR Title is: $PR_TITLE"` のままであり、タイトル全体はプロセスの環境変数（データ領域）に隔離された。
- シェルはダブルクォート内の環境変数を展開する際、記号をコードとして再解釈せず単なるひとつの引数（データ）として扱うため、インジェクションが完全に防止された。

#### 分かったこと
- PRタイトル、Issue本文、コミットメッセージ、ブランチ名など、**外部から入力可能なContext値は絶対に `run:` 内に直接 `${{ ... }}` で埋め込んではならない**。
- 必ず Step の `env:` にマッピングし、シェルスクリプト側では `"$VAR"` と参照するのがGitHub Actionsにおける必須のセキュリティベストプラクティスである。


## つまずいた点

- **`${{ ... }}` を文字列として書くと構文エラーになる**:
  - `echo "=== 1. Expression展開 (${{ ... }}) ==="` と記述したところ、GitHub Actionsのパーサーが `${{ ... }}` をExpression式として評価しようとし、`(Line: 22, Col: 14): Unexpected symbol: '...'. Located at position 1 within expression: ...` で即座にエラーとなった。
  - ワークフローファイル内でリテラルとして `${{` を出力したい場合は `${{ '${{ ... }}' }}` のようにエスケープするか、記号を避ける必要がある。
- **構文エラーがある場合のWorkflow Runの挙動**:
  - `push.branches: [main]` で作業ブランチへのpushが除外されている設定であっても、ワークフロー定義に構文エラーがあると、ブランチ判定以前にGitHubが即座に failure（Workflow file issue）のRunを作成する。
- **Expression式内の文字列クォート**:
  - シェルと異なり、GitHub ActionsのExpression言語内での文字列リテラルはシングルクォート `'...'` のみ。


## ブログへ残したい要点

- **Context/Expressionと環境変数は別物**:
  - Expression (`${{ ... }}`) はRunnerがシェルを呼ぶ前に静的に置換される（ハードコードと同義）。
  - 環境変数 (`$VAR`) はRunnerのOSプロセスに渡され、実行時にシェル自身が展開する。
- **スコープとシャドーイング**:
  - `env:` は Workflow > Job > Step の順で定義でき、より狭いスコープ（Step）が外側を上書きする。
  - Step単位の `env:` は他のStepへ漏れない（後続Stepに変数を渡すには `$GITHUB_ENV` 環境ファイルが必要）。
- **`if:` 条件式の書き方**:
  - `${{ ... }}` は省略可能で、省略するのが公式推奨スタイル。文字列リテラルはシングルクォートを使う。
- **セキュリティ（スクリプトインジェクションの防止）**:
  - 外部入力となり得るContext（PRタイトルやブランチ名など）を `run:` に直接 `${{ ... }}` で展開すると、悪意あるコマンドを実行される致命的な脆弱性になる。
  - 対策は至極明快で、**「必ず `env:` を経由させてシェルからは `"$VAR"` で参照する」** を徹底すること。


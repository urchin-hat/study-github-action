# Chapter 03: Jobの依存関係と並列実行

対応Issue: [#6 Jobの依存関係と並列実行を学ぶ](https://github.com/urchin-hat/study-github-action/issues/6)

## この章で理解したいこと

- 複数Jobのデフォルトの実行順序（並列実行）
- `needs` によるJob間の依存関係定義と直列実行
- GitLab CI/CDの `stages` とGitHub Actionsの `needs`（DAG: 有向非巡回グラフ）の違い
- 前段Jobから後続Jobへデータを渡す方法（Job Outputs）
- 前段Job失敗時の後続Jobの挙動と、実行制御（`always()`, `failure()`）
- Jobレベルのタイムアウト（`timeout-minutes`）とエラー無視（`continue-on-error`）

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 観点 |
| --- | --- | --- |
| `stages` / `stage` | 直接の対応なし（既定は並列実行） | Jobの実行順序・フェーズの定義 |
| `needs`（DAG機能） | `needs`（標準の依存関係機能） | 特定の先行Job完了を待つ条件 |
| `when: on_failure` / `when: always` | `if: failure()` / `if: always()` | 先行処理が失敗した場合の実行制御 |
| `artifacts:reports:dotenv` | `jobs.<job_id>.outputs` | Job間で小さな文字列データを渡す |
| `timeout` | `timeout-minutes` | Jobの最大実行時間制限 |
| `allow_failure: true` | `continue-on-error: true` | Jobの失敗を許容して後続を継続する |

GitLab CI/CDでは `stages: [build, test, deploy]` のようにパイプライン全体の「フェーズ」を明示的に定義し、各Jobがどのstageに属するかを指定する（ステージ内は並列、ステージ間は直列）。
一方、GitHub Actionsにはステージという概念自体がなく、**全Jobがデフォルトで並列に起動**し、順序を作りたい場合のみ **`needs` でJob同士を個別に結線（DAG: 有向非巡回グラフ）** する。

## 最初の方針

1. `needs` を書かない複数のJobを定義し、並列実行されることを確認する
2. `needs` を追加して、Jobの直列実行（依存関係グラフ）を確認する
3. `outputs` を定義して、前段Jobの値を後続Jobで受け取ってみる
4. 前段Jobを意図的に失敗させ、後続Jobのスキップ挙動および `always()` / `failure()` による救済を確認する
5. `timeout-minutes` や `continue-on-error` の挙動を確認する

## 実装前の予想

- [x] `needs` を指定しない場合、Jobは上から順番に動くか、それとも同時に動くか？: 予想は直列（上から順）。実際は同時に並列実行される。
- [x] GitLab CI/CDの「ステージ」のように、Jobをまとめる構文はGitHub Actionsにあるか？: 予想はある。実際は存在せず、`needs` によるDAG（依存グラフ）で順序を制御する。
- [x] 前段Jobが失敗した場合、`needs` で依存している後続Jobはどうなるか？: 予想はスキップ（B）。実際もスキップされる（暗黙の if: success() のため）。救済するには if: always() や if: failure() を使う。
- [x] 前段JobのStepでセットした環境変数やファイルは、`needs` で接続した別Jobへそのまま引き継がれるか？: 予想は同じRunnerで動くため見れる。実際はJobごとに毎回新しいRunnerが割り当てられるため引き継がれない（Job OutputsやArtifactが必要）。
- [x] 未指定時のタイムアウト上限は何時間か？: 予想は6時間（C）。実際も360分（6時間）。
- [x] `continue-on-error: true` のJobが失敗した場合、ワークフロー全体はどうなるか？: 予想は成功（緑）になる。実際も後続Jobは継続し、全体も成功扱いになる。


## 壁打ちメモ

### 2026-09-24: Chapter 03開始

Section 02のPR #15が`main`へmergeされたことを受け、`main`から`lesson/03-job-design`を作成した。

### 2026-09-24: Jobの実行順序とステージの有無を予想

1. デフォルトの実行順序:
   - 予想: 直列で動く（YAMLの上から順）。
   - 実際: **すべて並列（同時）で動く**。指定がなければ全Jobが同時にRunnerへ割り当てられる。
2. ステージ（Stage）の概念:
   - 予想: ある。
   - 実際: **存在しない**。GitLab CI/CDのような `stages` 構文はなく、Job単位で `needs` を使って個別に依存関係を結ぶ（DAGモデル）。

### 2026-09-24: `needs` による直列化とデータの共有を予想

1. 実行順序:
   - 予想: `job-a` の完了後に `job-b` が動く。
   - 実際: そのとおり。`needs: [job-a]` を指定することで、前段Jobの正常終了を待ってから後続Jobが起動する。
2. データの共有（環境変数・ファイル）:
   - 予想: 同じRunnerマシンで動くため見れる。
   - 実際: **Jobごとにまったく新しいクリーンなRunnerマシンが割り当てられる**ため、ファイルも環境変数も引き継がれない。Job間でデータを渡すには `outputs`（文字列）や `artifacts`（ファイル）を明示的に使う必要がある。

### 2026-09-24: 前段Job失敗時の後続Jobの挙動と救済方法を予想

1. 前段Job失敗時の後続Job:
   - 予想: B（スキップされる）。
   - 実際: そのとおり。後続Jobは実行されずに **スキップ（skipped）** される。デフォルトで各Jobには暗黙の `if: success()` が適用されているため。
2. 失敗時の救済・常時実行（通知など）:
   - 予想: `needs:` とは違う表現がある？
   - 実際: 依存関係は `needs:` のままにし、Jobレベルの **`if:`** にステータスチェック関数（`always()` や `failure()`）を指定する。
     - `if: always()`: 前段の成否に関わらず必ず実行（GitLabの `when: always` に相当）
     - `if: failure()`: 前段が失敗した場合のみ実行（GitLabの `when: on_failure` に相当）
     - 前段の結果は `${{ needs.<job_id>.result }}` で取得可能（`success` / `failure` / `skipped` / `cancelled`）

### 2026-09-24: ワークフローのメンタルモデル（DAGグラフとしての可読性）

- ユーザーの考察:
  - ワークフローは上から順に手続き的に実行されて途中で止まるスクリプトではない。
  - ワークフロー全体が評価され、依存関係（`needs`）と条件（`if`）に基づいたグラフ（DAG: 有向非巡回グラフ）として各Jobが実行・スキップされる。
  - そのため、コードを読む際も「上から下への手続き」ではなく「Job同士の結線図」として読む必要がある。

### 2026-09-24: `timeout-minutes` と `continue-on-error` を予想

1. タイムアウト上限（`timeout-minutes`）:
   - 予想: C（6時間 / 360分）。
   - 実際: そのとおり。GitHub-hosted Runnerのデフォルトタイムアウトは最大6時間。ハングしたジョブが枠を大量消費するのを防ぐため、実務では10〜30分などの明示設定が必須。
2. エラー許容（`continue-on-error`）:
   - 予想: ワークフロー全体は成功（緑）になる。
   - 実際: そのとおり。GitLab CI/CDの `allow_failure: true` に相当し、Jobが失敗しても後続Jobは通常どおり実行され、ワークフロー全体も成功（Success）として完了する。





## 試したことと結果

### 実験1: `needs` なしの複数Job並列実行

- PR: [#16](https://github.com/urchin-hat/study-github-action/pull/16)
- Workflow Run: [36011500629](https://github.com/urchin-hat/study-github-action/actions/runs/36011500629)

#### 実行結果
- **job-a**:
  - 開始: `14:15:48`
  - 終了: `14:15:51`
  - Runner Worker ID: `{106e2e9a-c8c0-4e2f-9677-be2a2895c323}` (Azure Region: westus)
- **job-b**:
  - 開始: `14:15:48`
  - 終了: `14:15:51`
  - Runner Worker ID: `{dccaba83-76b3-4c63-80b6-bf943627aec1}` (Azure Region: eastus)

#### 分かったこと
- **デフォルトは完全な並列実行**:
  - `job-a` と `job-b` はまったく同じ秒（14:15:48）に開始された。直列（上から順）ではなく、GitHub Actionsは定義されたJobを同時にRunnerへ割り当てて並列に実行する。
- **独立したRunner環境**:
  - Worker IDおよびRegion（westus と eastus）が異なっており、Jobごとにまったく別の独立したマシン（Runner）が立ち上がって並列稼働していることが確認できた。


### 実験2: `needs` による直列実行と Job Outputs によるデータ受け渡し

- PR: [#16](https://github.com/urchin-hat/study-github-action/pull/16)
- Workflow Run: [36012058315](https://github.com/urchin-hat/study-github-action/actions/runs/36012058315)

#### 実行結果
- **job-a**:
  - 開始: `14:20:22` / 終了: `14:20:24`
  - Worker ID: `{b7e5c959-...}` (Azure Region: northcentralus)
  - `$GITHUB_OUTPUT` に `token=secret-token-xyz` を出力し、`outputs.my_token` で外部公開
- **job-b**:
  - 開始: `14:20:29`（`job-a` 終了後に起動）
  - Worker ID: `{436d6556-...}` (Azure Region: eastus)
  - ログ出力：
    ```text
    === 1. ファイルの確認 ===
    ls: cannot access 'job_a_file.txt': No such file or directory
    File not found! (別マシンなので存在しない)
    === 2. 環境変数の確認 ===
    MY_VAR: '' (別マシンなので空)
    === 3. Job Outputsの確認 ===
    Token from job-a: secret-token-xyz
    ```

#### 分かったこと
- **直列化の成功**: `needs: [job-a]` を指定したことで、`job-a` 完了（14:20:24）後に `job-b` が起動（14:20:29）した。
- **直列でもRunnerマシンは別物（完全な隔離）**:
  - 「直列で動くなら同じマシンで連続実行される」という予想に反し、`job-b` にはまったく別のクリーンなRunner（Regionも別）が割り当てられた。
  - そのため、前段Jobで作成したファイルや `$GITHUB_ENV` で設定した環境変数は、後続Jobには一切引き継がれない。
- **Job Outputs による明示的な受け渡し**:
  - Job間で文字列データを渡すには、`$GITHUB_OUTPUT` に書き込み、Job定義の `outputs:` で公開し、後続Jobから `${{ needs.<job_id>.outputs.<name> }}` で参照する必要がある。

### 実験3: 前段Job失敗時のスキップ挙動と `always()` による救済

- PR: [#16](https://github.com/urchin-hat/study-github-action/pull/16)
- Workflow Run: [36012847241](https://github.com/urchin-hat/study-github-action/actions/runs/36012847241)

#### 実行結果
各Jobのステータス：
- `test-job`: ❌ **failure**（`exit 1` により2秒で異常終了）
- `deploy-job`: ⚪ **skipped**（実行されずスキップ）
- `notify-job`: ✅ **success**（前段が失敗しても実行された！）

`notify-job` のログ出力：
```text
=== Notification Service ===
test-job result was: failure
Alert: Deployment was prevented because test-job failed.
```

#### 分かったこと
- **デフォルトはスキップ**: `needs: [test-job]` のみで `if:` を書かないJob（`deploy-job`）は、暗黙の `if: success()` により、前段が失敗した時点で **一切実行されずにスキップ（0秒）** される。本番デプロイなど後続を安全に止めたい場合はデフォルトのままで良い。
- **`always()` による救済**: 通知のように前段が失敗しても動かしたいJobには `if: always()` を明示すれば、パイプラインが途中で打ち切られることなく確実に実行される。
- **前段のステータス参照**: `${{ needs.<job_id>.result }}` で前段Jobの終了状態（`failure` など）を文字列として参照できる。

### 実験4: `timeout-minutes` と `continue-on-error`

- PR: [#16](https://github.com/urchin-hat/study-github-action/pull/16)
- Workflow Run: [36013707798](https://github.com/urchin-hat/study-github-action/actions/runs/36013707798)

#### 実行結果
- `flaky-job`: ❌ **failure**（`exit 1` で失敗したが `continue-on-error: true` を設定）
- `downstream-job`: ✅ **success**（`needs: [flaky-job]` のみだがスキップされずに正常実行！）
- ワークフロー全体の結果: ✅ **success（緑のチェック）**

`downstream-job` のログ出力：
```text
Downstream job executed successfully because flaky-job was allowed to fail!
```

#### 分かったこと
- **`continue-on-error: true` によるエラー許容**:
  - GitLab CI/CDの `allow_failure: true` と同等に機能し、Job自身が失敗しても後続Jobは通常通り実行される。
  - ワークフロー全体の結果も失敗（Failure）にはならず、成功（Success）として完了する。
- **タイムアウト設定の重要性**:
  - デフォルトのタイムアウトは6時間（360分）と非常に長いため、ハング時のリソース・料金消費を防ぐために `timeout-minutes: 5` などの明示設定が不可欠である。


## つまずいた点

- **直列接続（`needs`）でも環境は引き継がれない**:
  - `needs` で接続すれば同じ環境で継続実行されると誤解しやすいが、GitHub ActionsではJob単位で毎回クリーンな新しい仮想マシン/コンテナが起動する。
  - ファイルを共有したい場合はArtifacts、変数を共有したい場合はJob Outputsが必須となる。
- **ステージ概念が存在しないことのギャップ**:
  - GitLab CI/CDのように `stages` を定義して「フェーズ」で束ねるのではなく、Job個々の結線（`needs`）によるDAGモデルで設計する必要がある。


## ブログへ残したい要点

- **ワークフローのメンタルモデル（手続き型ではなくDAGグラフ）**:
  - ワークフローは上から下へ流れるスクリプトではなく、Job同士の依存関係（`needs`）と条件（`if`）を定義したグラフ（DAG: 有向非巡回グラフ）として評価される。
  - 依存がなければすべて並列実行され、どこかが失敗してもパイプライン全体が即座に打ち切られるわけではない。
- **直列化してもJob間は完全隔離**:
  - `needs` で順番を作っても、Jobごとに別々のRunnerマシンが起動する。環境変数やファイルは共有されない。
  - 文字列データの受け渡しには `$GITHUB_OUTPUT` と `outputs` を使う。
- **後続Jobの制御**:
  - デフォルトでは暗黙の `if: success()` により、前段がコケると後続は安全にスキップされる（デプロイ防止）。
  - 失敗時でも動かしたい通知処理等には `if: always()` を明示し、`${{ needs.<job_id>.result }}` で前段の成否を拾う。
- **タイムアウトとエラー許容**:
  - デフォルトタイムアウト（6時間）の罠を避けるため、`timeout-minutes` の設定を習慣化する。
  - 許容可能なテスト失敗には `continue-on-error: true` を活用する。


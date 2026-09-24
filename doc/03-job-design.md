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
- [ ] 前段Jobが失敗した場合、`needs` で依存している後続Jobはどうなるか？
- [ ] 前段JobのStepでセットした環境変数は、`needs` で接続した別Jobへそのまま引き継がれるか？

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


## つまずいた点

壁打ちと実装の進行に合わせて追記する。

## ブログへ残したい要点

壁打ちと実装の進行に合わせて追記する。

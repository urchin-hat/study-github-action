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

- [ ] `needs` を指定しない場合、Jobは上から順番に動くか、それとも同時に動くか？
- [ ] GitLab CI/CDの「ステージ」のように、Jobをまとめる構文はGitHub Actionsにあるか？
- [ ] 前段Jobが失敗した場合、`needs` で依存している後続Jobはどうなるか？
- [ ] 前段JobのStepでセットした環境変数は、`needs` で接続した別Jobへそのまま引き継がれるか？

## 壁打ちメモ

### 2026-09-24: Chapter 03開始

Section 02のPR #15が`main`へmergeされたことを受け、`main`から`lesson/03-job-design`を作成した。

## 試したことと結果

壁打ちと実装の進行に合わせて追記する。

## つまずいた点

壁打ちと実装の進行に合わせて追記する。

## ブログへ残したい要点

壁打ちと実装の進行に合わせて追記する。

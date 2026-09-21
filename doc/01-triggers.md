# Chapter 01: WorkflowのTrigger

対応Issue: [#4 WorkflowのTriggerを学ぶ](https://github.com/urchin-hat/study-github-action/issues/4)

## この章で理解したいこと

- `push`、`pull_request`、`workflow_dispatch`の違い
- 1つのworkflowを複数eventから起動する方法
- branch filterとpath filterの役割
- workflowが起動しなかった場合の切り分け方

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 観点 |
| --- | --- | --- |
| Pipeline source | Event | 何が処理を起動したか |
| `workflow: rules` | `on` | Workflow自体を作成する条件 |
| Jobの`rules` | JobやStepの`if` | 作成されたWorkflow内で処理を実行する条件 |
| branch/path条件 | `branches` / `paths` | 対象となる変更を絞り込む |

`on`と`if`を混同しないことがこの章の重要点になる。`on`に一致しなければworkflow run自体が
作られない。一方、`if`は作られたrunの中でjobまたはstepを実行するか判断する。

## Chapter 00から分かったこと

Chapter 00のworkflowは`on: push`だけを指定した。そのため、workflowファイルだけでなく
Chapter 00のメモだけをpushしたときにもrunが作成された。path filterを指定しなければ、
変更ファイルの種類に関係なくpush eventへ反応する。

## 最初の方針

既存のHello World workflowを使い、アプリの処理は増やさず起動条件だけを変更する。

1. 複数eventを指定する
2. event名をログへ表示して違いを確認する
3. branch filterを追加して一致・不一致を試す
4. path filterを追加して一致・不一致を試す
5. 手動実行を試す

## 実装前の予想

- [ ] 複数eventは`on`へどのようなYAMLで指定するか
- [ ] Pull Requestの作成・更新を表すevent名は何か
- [ ] 手動実行を有効にするevent名は何か
- [ ] event名を実行ログへ表示するには何を参照するか
- [ ] branch filterとpath filterを両方指定した場合、ORとANDのどちらになるか

## 壁打ちメモ

### 2026-09-21: Chapter 01開始

PR #12と追加メモを含む変更が`main`へmergeされた後、最新の`main`から
`lesson/01-triggers`を作成した。

## 試したことと結果

壁打ちと実装の進行に合わせて追記する。

## つまずいた点

壁打ちと実装の進行に合わせて追記する。

## ブログへ残したい要点

- Triggerは「いつWorkflowを作るか」を決める
- Triggerを学ぶ間は実行処理を単純なままにし、起動条件の差だけを観察する

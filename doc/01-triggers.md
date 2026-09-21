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

### 2026-09-21: 複数eventの指定方法を予想

`push`、Pull Request、手動実行をまとめて指定する方法として`all`を予想した。
GitHub Actionsにすべてのeventを意味する`all`指定はなく、受け取りたいeventを明示的に列挙する。

今回必要なevent名は次の3つ。

- `push`
- `pull_request`
- `workflow_dispatch`

filterやeventごとの設定が不要な段階では、`on`の値を配列にする短縮形で指定できる。

複数eventを「`push`や`pull_request`ごとのセクションに分ける」と捉えた。考え方は近いが、
正確には`on`の中に複数のeventを列挙する。設定がなければ配列の短縮形、eventごとにfilterを
設定する場合はmapping形式を使う。

```yaml
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:
```

この例の`push`、`pull_request`、`workflow_dispatch`は別々のWorkflowではなく、同じWorkflowを
開始できる3種類のeventである。

### 2026-09-21: Workflowファイル名を検討

Chapter 01用にどのファイル名がよいかを検討した。今回は新しいWorkflowを追加するのではなく、
Chapter 00で作った同じWorkflowのTriggerを変更するため、`.github/workflows/hello-world.yml`を
そのまま使う。

`.github/workflows/`内のファイル名は管理者が用途に合わせて決められる。Actions画面へ表示される
Workflow名はファイル名ではなくYAML内の`name`で決まる。実務では、例えばCIなら`ci.yml`、
deployなら`deploy.yml`のように役割で分けると読みやすい。

学習章ごとにWorkflowファイルを追加すると、過去のWorkflowも`push`へ反応し、同じpushで複数の
runが起動する。その違いを学ぶ目的がない限り、今回はファイルを増やさない。

### 2026-09-21: 3つのeventとbranch filterを設定

既存の`hello-world.yml`へ次のTriggerを設定した。

```yaml
on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main
  workflow_dispatch:
```

配列の短縮形ではなくmapping形式を使い、`push`と`pull_request`へそれぞれ`main`のbranch
filterを設定した。`workflow_dispatch`は追加設定がないため、値が空でも有効である。

この設定をpushする前に、現在の`lesson/01-triggers`へのpushでrunが作られるかを予想する。

予想は「`main`を指定しているため、作業branchへのpushでも`main`向けPRでも起動しない」だった。
正しくは次のようになる。

- `push.branches`: pushされたbranchを判定する。`lesson/01-triggers`へのpushは`main`ではないため起動しない
- `pull_request.branches`: PRの取り込み先であるbase branchを判定する。`main`向けPRなので起動する

同じ`branches: [main]`でも、eventによって比較対象が異なる点に注意する。

`workflow_dispatch`はpushやPRで自動起動する条件ではなく、手動起動を許可するeventである。
GitHub UIに`Run workflow`ボタンを表示するには、`workflow_dispatch`を含むWorkflowファイルが
default branchに存在する必要がある。そのため、この変更を`main`へmergeする前はUIからの手動
実行をまだ試せない。

### 2026-09-21: branch filterに一致しないpushを実行

commit `3b80f6c`を`lesson/01-triggers`へpushした。`push.branches`は`main`だけを対象としている
ため、このcommitに対するworkflow runとcheck runはどちらも0件だった。

これはworkflowが起動して失敗またはskipした状態ではない。Trigger条件に一致せず、workflow
run自体が作成されなかった状態である。Actionsの問題を調査するときは、次を区別する必要がある。

- runが存在しない: `on`やfilter、workflowファイルの配置を確認する
- run内のjobがskip: `if`や`needs`を確認する
- jobまたはstepがfailure: 実行ログとexit codeを確認する

### 2026-09-21: EventとGit refをログへ追加

次の2つを表示するようにstepを変更した。

```yaml
run: |
  echo "event: ${{ github.event_name }}"
  echo "branch: ${{ github.ref }}"
```

依頼した`github.event_name`に加え、`github.ref`も追加した。どちらもGitHubがWorkflowを
評価するときにexpressionを値へ置換し、その結果をshell commandへ渡す。

一方、stepはHello Worldを表示しなくなったため、表示名`Hello, World`と実際の処理が一致しなく
なった。Actionsのログから役割を判断できるよう、step名も処理内容に合わせて変更する。

## 試したことと結果

壁打ちと実装の進行に合わせて追記する。

## つまずいた点

壁打ちと実装の進行に合わせて追記する。

## ブログへ残したい要点

- Triggerは「いつWorkflowを作るか」を決める
- Triggerを学ぶ間は実行処理を単純なままにし、起動条件の差だけを観察する

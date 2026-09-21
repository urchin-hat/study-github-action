# Chapter 00: Hello World

対応Issue: [#3 Hello World workflow](https://github.com/urchin-hat/study-github-action/issues/3)

## この章で理解したいこと

- workflow、job、step、Runnerの関係
- workflowファイルを置く場所
- pushからworkflow実行までの流れ
- Actions画面で実行結果とログを確認する方法

## GitLab CI/CDとの対応

学習開始時点では、次のように対応づけて考えます。実際に試して認識が変わった場合は、
この表へ追記します。

| GitLab CI/CD | GitHub Actions | 現時点の理解 |
| --- | --- | --- |
| Pipeline | Workflow | eventを契機に開始される一連の自動処理 |
| Job | Job | Runner上で実行される処理単位 |
| `script` | Stepの`run` | shell commandを実行する |
| Runner | Runner | jobを実際に実行する環境 |

GitLab CI/CDのstageに相当するものは、この章ではまだ扱いません。

## 最初の方針

- サンプルアプリはまだ作らない
- repositoryのcheckoutもまだ行わない
- 外部Actionはまだ利用しない
- GitHub-hosted Ubuntu Runnerで`Hello, world!`を表示するだけにする
- 最小構成を成功させた後、意図的に失敗させてログの違いを見る

## 壁打ちメモ

### 2026-09-21: 学習開始

最初からGoアプリのtestやbuildを題材にすると、GitHub Actionsそのものとアプリ固有の処理を
同時に考えることになる。そのため、最初の章は`echo`だけを使い、workflowの構造と実行画面の
理解へ集中することにした。

学習はセクションごとにbranchとPRを分ける。Chapter 00は最新の`main`から作成した
`lesson/00-hello-world`で進める。

### 2026-09-21: 最小構成のキーを予想

最初の予想:

1. pushの指定は`on`
2. Runnerの指定は`run-on`
3. `echo` commandを書く場所は`run`

1と3は正解。2も役割の理解は合っているが、正しいキー名は複数形の`runs-on`だった。
`run`はstepが実行するcommandを表し、`runs-on`はjobを実行するRunnerを表す。

`runs-on`はstepではなくjobに設定する。同じjob内のstepは、そこで選ばれた同じRunner上で
上から順番に実行されるためである。

### 2026-09-21: 編集用のworkflowを作成

`.github/workflows/hello-world.yml`に値が空の骨組みを作成した。`name`、`on`、`jobs`は
同じトップレベルのキーであるため、同じ深さに揃えた。ここから値を1つずつ埋める。

### 2026-09-21: 最初のYAMLを記述

最初に次の値を選んだ。

- workflow名: `chap1`
- event: `push`
- Runner: `ubuntu-latest`
- command: `echo "hello,world"`

workflow、event、Runner、commandの置き場所は合っていた。一方、`name:"chap1"`のように
コロンの直後へ値を書いており、YAMLのkey/valueとして解釈させるために必要な空白がなかった。
`name: "chap1"`または`name: chap1`のように、コロンの後へ空白を入れる必要がある。

また、stepの`name`が空だった。step名は処理自体には使われない表示用の名前だが、Actionsの
実行ログを読みやすくするために具体的な名前を付ける。

### 2026-09-21: YAMLを修正

コロン後の空白を追加し、step名を`Hello, World`にした。完成した最小workflowは次の構造に
なった。

- workflowの表示名: `chap1`
- 起動event: `push`
- job ID: `hello`
- Runner: `ubuntu-latest`
- stepの表示名: `Hello, World`
- 実行command: `echo "hello,world"`

すべての値を引用符で囲んでいるが、この書き方は有効である。単純な文字列では引用符を省略
する書き方もできるため、今後は可読性を見ながら使い分ける。

### 2026-09-21: checkoutなしの実行結果を予想

実行前の予想は「repositoryをcheckoutしていないため成功しない」だった。

今回のcommandはRunnerに最初から存在するshellの`echo`だけを使い、repository内のファイルを
参照しない。そのため、checkoutしていなくても実行できるはずである。GitHub Actionsでは
Runnerの準備とrepositoryのcheckoutは別の処理であり、checkoutが必要なのはソースコードや
repository内のscriptを利用するときである。

この予想の違いを、実際にpushして確認する。

## 実装前の予想

workflowを書く前に、次の問いへの予想を記録する。

- [x] workflowファイルはrepository内のどこへ置くか: `.github/workflows/`
- [x] pushされたことをYAMLのどの項目で指定するか: `on`
- [x] jobを動かすOSをどこで指定するか: jobの`runs-on`
- [x] `Hello, world!`を実行するcommandをどこへ書くか: stepの`run`
- [x] repositoryのcheckoutなしで`echo`を実行できるか: 実行できる

## 試したことと結果

`lesson/00-hello-world`をpushし、workflow run
[35585930288](https://github.com/urchin-hat/study-github-action/actions/runs/35585930288)を実行した。

結果は成功。`actions/checkout`を使っていなくても、Runner上のshellで`echo`が実行され、
ログに`hello,world`と表示された。

実行ログから次のことも確認できた。

- `ubuntu-latest`としてUbuntu 24.04のGitHub-hosted Runnerが割り当てられた
- 記述したstepの前後に`Set up job`と`Complete job`が自動的に実行された
- `run`は既定で`/usr/bin/bash -e`を使って実行された
- job IDの`hello`とstep名の`Hello, World`がログ上の異なる階層に表示された
- repositoryをcheckoutするstepは存在しなかった

この結果から、Runnerを用意することとrepositoryの内容をRunnerへ取得することは別の処理だと
確認できた。

### 2026-09-21: 意図的な失敗方法を検討

失敗させる方法として、Ubuntu Runnerに存在しないcommandの実行と`exit 1`を予想した。
どちらも非0のexit codeになるため、stepは失敗する。

- 存在しないcommand: shellでは通常exit code 127になるが、名前の間違いなのか意図した失敗かが曖昧
- `exit 1`: 意図した場所で明示的に失敗させられ、結果を予想しやすい

今回は、まずメッセージを出力してから`exit 1`を実行する。複数行のcommandを書くため、YAMLの
block scalarである`|`を使う。

最初の修正では`run: | `のように`|`の後ろへ空白が残り、`git diff --check`でtrailing
whitespaceとして検出された。workflowの意味とは直接関係しない空白でも、レビュー時の不要な
差分やlintエラーを避けるため削除する。

workflow run
[35586273665](https://github.com/urchin-hat/study-github-action/actions/runs/35586273665)で
意図的な失敗を実行した。結果は予想どおりfailureになり、ログには次の順序で表示された。

```text
Run echo "hello,world"
hello,world
Error: Process completed with exit code 1.
```

`echo`が成功しても、その後の`exit 1`が非0を返したためstep全体が失敗した。必須stepの失敗は
jobの失敗となり、このworkflowにはjobが1つだけなのでworkflow全体も失敗した。

その後`exit 1`を削除して再度pushし、workflow run
[35586441031](https://github.com/urchin-hat/study-github-action/actions/runs/35586441031)が
successへ戻ることを確認した。失敗原因をログで特定し、その原因だけを修正して成功状態へ戻す
一連の流れを体験できた。

### 2026-09-21: 実行単位を自分の言葉で整理

学習後、4つの用語を次のように説明した。

1. Workflow: プロセスの基本単位
2. Job: Workflow内の処理
3. Runner: Workflowを動かすOS
4. Step: 各プロセスの単位

Workflow、Job、Stepが外側から内側へ並ぶ理解は合っていた。より正確には次のように整理できる。

- Workflow: eventを契機に実行される、自動化全体の定義
- Job: Workflowを構成し、1台のRunner上で実行される処理単位
- Runner: Jobを実行するマシンまたは実行環境。OSそのものではない
- Step: Job内で上から順に実行される最小の処理単位。commandまたはActionを実行する

`runs-on: ubuntu-latest`はRunnerのlabelを指定しており、その条件に合うUbuntu環境がjobへ
割り当てられる。今回の構成は`chap1` Workflowの中に`hello` Jobがあり、そのJobがUbuntu
Runner上で`Hello, World` Stepを実行する、という関係になる。

## 補足: Slackで成功・失敗を検知する

方法は大きく2つある。

### GitHubのSlack Appで購読する

通知を見ることが目的なら最も簡単な方法。Slack workspaceへGitHub Appを導入し、通知先の
channelで次のように購読する。

```text
/github subscribe urchin-hat/study-github-action workflows:{name:"chap1" event:"push"}
```

workflowの開始時に通知され、完了すると同じthreadが成功または失敗の結果で更新される。
name、event、branch、actorでfilterできる。workflowへ通知stepやSlack credentialを追加せずに
済むため、単純な監視ではこの方法から検討する。

公式資料: [Customizing notifications for GitHub in Slack](https://docs.github.com/en/integrations/how-tos/slack/customize-notifications)

### WorkflowからSlackへ送信する

通知文面、mention、送信条件を細かく制御したい場合は、Slack Incoming Webhookなどを使う。
Webhook URLはrepository secretの`SLACK_WEBHOOK_URL`へ保存し、YAMLへ直接書かない。

複数jobの最終結果を通知する場合の概念的な構成は次のようになる。

```yaml
notify:
  needs: [hello]
  if: ${{ always() }}
  runs-on: ubuntu-latest
  steps:
    - name: Notify Slack
      uses: slackapi/slack-github-action@<FULL_COMMIT_SHA>
      with:
        webhook: ${{ secrets.SLACK_WEBHOOK_URL }}
        webhook-type: incoming-webhook
        payload: |
          text: "chap1: ${{ needs.hello.result }}"
```

通常、失敗したjobに依存する後続jobはskipされる。通知jobへ`if: ${{ always() }}`を付けることで、
前段が失敗しても通知を試行できる。通知job自身の`job.status`ではなく、
`needs.hello.result`で前段jobの結果を参照する点にも注意する。

第三者Actionはtagだけでなく完全なcommit SHAへ固定すると、参照先を書き換えられるリスクを
下げられる。実装する場合はSecurityの章でsecret、permissions、Actionの固定方法と合わせて扱う。

## つまずいた点

- YAMLではmappingのコロンと値の間に空白が必要
- workflowの`name`とstepの`name`は別のもので、Actions画面では異なる階層に表示される

## ブログへ残したい要点

- CI/CD経験者でも、最初はアプリのbuildから入らず最小workflowの実行単位を確認するとよい
- GitLab CI/CDとの名称の対応だけでなく、実行モデルの違いを実験で確認する

# GitHub Actions 1日学習カリキュラム

GitLab CI/CDを使った経験がある人向けに、GitHub Actionsの基本を1日で段階的に学ぶ
カリキュラムです。

このPRではカリキュラムだけを決めます。サンプルアプリやworkflowはまだ実装しません。

## 学習のゴール

1日の終了時に、次のことができる状態を目指します。

- workflow、job、step、Action、Runnerの関係を説明できる
- push、Pull Request、手動操作を起点にworkflowを実行できる
- jobの依存関係、並列実行、条件分岐を設定できる
- cacheとartifactを使い分けられる
- secretsと`GITHUB_TOKEN`を最小権限で扱える
- Actions画面のログから失敗原因を調査できる

## 学習の進め方

セクションごとに新しいbranchを作り、Pull Requestで`main`へmergeします。次のbranchは、
直前のPRをmergeした後の`main`から作ります。

```text
main
  └─ lesson/00-hello-world ── PR・merge ── main
                                               └─ lesson/01-triggers ── PR・merge
                                                                              └─ ...
```

ルール:

- `main`へ直接commit・pushしない
- 1つのbranchでは1つの学習テーマだけを扱う
- branch名は`lesson/<番号>-<テーマ>`とする
- 各PRに「変更内容」「確認方法」「学んだこと」を記録する
- workflowが失敗した状態ではmergeしない
- mergeにはSquash mergeを使い、merge後のbranchは削除する

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 主な違い |
| --- | --- | --- |
| Pipeline | Workflow | `.github/workflows/*.yml`ごとに定義する |
| `workflow: rules` | `on` | event、branch、pathで起動条件を指定する |
| Stage | 直接の対応なし | `needs`でjobの依存関係を作る |
| Job | Job | jobごとに独立したRunnerで実行される |
| `script` | `steps[].run` | shell commandを実行する |
| Runner tags | `runs-on` | RunnerのOSやlabelを指定する |
| Variables | `env`、`vars`、`secrets` | 機密情報は`secrets`を使う |
| Artifacts | Artifacts | job間のファイル受け渡しにも使う |
| Cache | Cache | 再生成可能な依存データの高速化に使う |
| `include` | Reusable workflow | 複数workflowで処理を共有する |
| `resource_group` | `concurrency` | 同じ対象への同時実行を制御する |

## GitHub Actionsの課金体系

以下は2026年9月21日時点のGitHub公式情報です。料金は変更される可能性があるため、実際に
有料利用を始める前に[GitHub Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions)
と[Runner料金表](https://docs.github.com/en/billing/reference/actions-runner-pricing)を再確認します。

### 何に料金がかかるか

主な課金対象は次の2つです。

1. GitHub-hosted Runnerを実行した時間
2. artifact、cache、custom imageを保存した容量と期間

standard GitHub-hosted Runnerはpublic repositoryでは無料です。private repositoryでは
planごとの無料枠を消費し、超過分が課金されます。self-hosted RunnerにGitHub Actionsの
実行料金はかかりませんが、サーバーやクラウドなどRunner自体の費用は利用者が負担します。

Larger Runnerはpublic repositoryでも常に有料で、planの無料minutesを利用できません。

この学習リポジトリ`urchin-hat/study-github-action`はpublicです。そのため、
`ubuntu-latest`などstandard GitHub-hosted Runnerの実行minutesは無料です。ただし、
artifactなどのstorage、Larger Runner、外部サービスやself-hosted Runnerのインフラ費用まで
すべて無料になるという意味ではありません。

| RepositoryとRunner | 実行minutesの扱い |
| --- | --- |
| Public + standard GitHub-hosted Runner | 無料 |
| Private + standard GitHub-hosted Runner | planの無料枠を消費し、超過分は有料 |
| Public/Private + Larger Runner | 常に有料。無料minutesは利用不可 |
| Public/Private + self-hosted Runner | GitHub Actionsの実行料金は無料。Runnerの運用費は自己負担 |

public repositoryではソースコード、workflow、実行ログも公開されます。料金面では学習しやすい
一方で、credential、個人情報、非公開URLなどをcommitやログへ出さないことが前提です。

### Planごとの無料枠

| Plan | private repositoryの実行時間/月 | Artifact storage | Cache storage |
| --- | ---: | ---: | ---: |
| GitHub Free | 2,000分 | 500 MB | repositoryごとに10 GB |
| GitHub Pro | 3,000分 | 1 GB | repositoryごとに10 GB |
| GitHub Free for organizations | 2,000分 | 500 MB | repositoryごとに10 GB |
| GitHub Team | 3,000分 | 2 GB | repositoryごとに10 GB |
| GitHub Enterprise Cloud | 50,000分 | 50 GB | repositoryごとに10 GB |

minutesはbilling cycleの開始時にリセットされ、workflowを起動した人ではなくrepositoryの
ownerへ計上されます。Artifact storageの枠はGitHub Packagesと共有されます。Cache storageは
artifactとは別枠です。

### 無料枠を超えた場合の基準料金

standard Runnerの主な料金は次のとおりです。

| Runner | 料金（USD/分） |
| --- | ---: |
| Linux x64 1-core | $0.002 |
| Linux x64 2-core | $0.006 |
| Linux arm64 2-core | $0.005 |
| Windows x64/arm64 2-core | $0.010 |
| macOS 3-coreまたは4-core | $0.062 |

課金時間はjobごとに1分単位へ切り上げられます。たとえば10秒のjobを6個に分けた場合、
合計実行時間が約1分でも、課金上は最大6分として扱われます。job分割は並列性や責務の分離と、
この切り上げの影響を両方考えて決めます。

追加storageの基準料金:

| Storage | 料金（USD/GB・月） |
| --- | ---: |
| ArtifactとGitHub Packagesの共有storage | $0.25 |
| Actions cache | $0.07 |
| Larger Runnerのcustom image | $0.07 |

storageは単純な月末容量ではなく、実際に保存していた時間をGB-hoursで積算します。途中で
artifactを削除すると、それ以降の加算は止まりますが、それまでに積算された利用量は消えません。

### 無料枠を使い切ったとき

- 有効な支払い方法がない場合、無料枠を使い切ると追加利用が停止する
- 支払い方法がある場合、設定したbudgetに応じて超過利用が課金される
- budgetは通知だけにも、上限到達時に利用を停止するhard limitにもできる
- included usageが90%と100%へ到達したときのemail通知を有効にできる

学習開始前に`Settings > Billing and licensing > Budgets and alerts`を確認し、意図しない
課金を避けたい場合はActions用budgetを作って`Stop usage when budget limit is reached`を
有効にします。

### このカリキュラムでの節約ルール

- standardの`ubuntu-latest`を基本にする
- macOS、Windows、Larger Runnerは必要な演習でだけ使う
- `timeout-minutes`を設定して無限待ちを防ぐ
- 同じcommitに対する不要なpushを減らす
- matrixを増やしすぎない
- artifactには短い`retention-days`を設定する
- cache対象をdependencyに限定し、build成果物全体をcacheしない
- 不要になったworkflow runとartifactを削除する
- public repositoryでも機密情報を置かない

artifactとlogの標準保持期間は90日です。この教材ではartifact演習時に短い保持期間を指定し、
保存期間がstorage使用量へどう影響するかも確認します。

## 1日の予定

| 時間 | セクション | 学ぶ内容 |
| --- | --- | --- |
| 09:00–09:30 | 00: Hello World | 最小workflowとActions画面 |
| 09:30–10:20 | 01: Trigger | push、Pull Request、手動実行 |
| 10:20–10:30 | 休憩 |  |
| 10:30–11:20 | 02: Contextと変数 | expressions、context、`env` |
| 11:20–12:10 | 03: Job設計 | 並列実行、`needs`、条件分岐 |
| 12:10–13:10 | 昼休憩 |  |
| 13:10–14:30 | 04: サンプルアプリCI | lint、test、build、matrix |
| 14:30–15:20 | 05: データの受け渡し | cacheとartifact |
| 15:20–15:30 | 休憩 |  |
| 15:30–16:20 | 06: セキュリティ | permissions、secrets、安全な入力 |
| 16:20–17:10 | 07: Deploy設計 | Environment、concurrency、手動deploy |
| 17:10–17:30 | 振り返り | 障害調査と理解度確認 |

## 00: Hello World

Branch: `lesson/00-hello-world`

最初はアプリ、checkout、変数、外部Actionを使いません。pushされたときに
GitHub-hosted Runner上で`Hello, world!`を表示するだけのworkflowを作ります。

学ぶ項目:

- `.github/workflows/`にYAMLを置く
- `name`、`on`、`jobs`、`runs-on`、`steps`、`run`
- workflow、job、stepの実行ログを見る
- 成功・失敗・再実行をActions画面で確認する

完成条件:

- branchをpushするとworkflowが自動実行される
- Actions画面で`Hello, world!`を確認できる
- YAMLの各行が何を表すか説明できる

## 01: Trigger

Branch: `lesson/01-triggers`

Hello World workflowの起動方法を増やします。

- `push`
- `pull_request`
- `workflow_dispatch`
- branch filterとpath filter

同じworkflowを複数eventから実行し、どのeventで開始されたかをログで確認します。

## 02: Contextと変数

Branch: `lesson/02-contexts-and-variables`

まだサンプルアプリは使わず、ログ出力だけで値の評価方法を確認します。

- `${{ github.event_name }}`などの`github` context
- `${{ ... }}`とshell環境変数の違い
- workflow、job、stepごとの`env` scope
- `if`によるstepの条件分岐
- 外部入力を安全にshellへ渡す方法

## 03: Job設計

Branch: `lesson/03-job-design`

複数jobを使い、GitLab CI/CDのstageとの違いを確認します。

- `needs`のないjobの並列実行
- `needs`を使った依存関係
- step outputとjob output
- `success()`、`failure()`、`always()`
- `timeout-minutes`と`continue-on-error`

ここまでは`echo`や簡単なshell commandだけで学びます。

## 04: サンプルアプリCI

Branch: `lesson/04-application-ci`

このセクションへ進む前に、使用する言語とサンプルアプリを決めます。候補は外部依存の
少ない小さなWebアプリです。決定後、次の処理を別々のjobとして組み立てます。

```text
lint ----\
          -> build
test ----/
```

- 公式のsetup Action
- dependencyの取得
- lintまたはformat check
- unit test
- build
- version matrix

## 05: CacheとArtifact

Branch: `lesson/05-cache-and-artifact`

- dependency cacheで2回目の実行を高速化する
- cache hitとcache missを確認する
- build成果物をartifactとして保存する
- 別jobでartifactをdownloadして検証する

cacheは再生成可能な高速化用データ、artifactは実行結果として残すファイル、という違いを
実際のworkflowで確認します。

## 06: セキュリティ

Branch: `lesson/06-security`

- `permissions`で`GITHUB_TOKEN`を最小権限にする
- `vars`と`secrets`を使い分ける
- forkからのPull Requestでsecretが渡らない理由を理解する
- 第三者Actionの発行元とversion指定を確認する
- 未信頼の入力をshellへ直接展開しない
- `pull_request_target`の危険な使用方法を知る

## 07: Deploy設計

Branch: `lesson/07-deploy`

実サービスには接続せず、artifactを取得してdeploy先を表示するところまでを作ります。

- `workflow_dispatch`の入力で`staging`か`production`を選ぶ
- GitHub Environmentをjobへ設定する
- `concurrency`で同じ環境への同時deployを防ぐ
- productionの承認ルールを検討する
- cloud接続時にOIDCを使う理由を理解する

## 各セクション共通の確認方法

1. 実行前に、どのjobとstepが動くか予想する。
2. branchをpushしてActions画面を確認する。
3. 実行グラフと各stepのログを見る。
4. 意図的に1か所壊し、どこで失敗するか確認する。
5. 修正して成功させる。
6. PRに学んだことを追記してmergeする。

## 初日の対象外

- 独自JavaScript ActionやDocker Actionの開発
- self-hosted Runnerの構築
- Organization全体のRunner管理
- 実クラウド環境への本番deploy
- 大規模なReusable workflow設計

これらは基本的なworkflowを自力で読んで修正できるようになってから扱います。

## 公式リファレンス

- [GitHub Actionsを理解する](https://docs.github.com/ja/actions/get-started/understand-github-actions)
- [Workflow構文](https://docs.github.com/ja/actions/reference/workflows-and-actions/workflow-syntax)
- [Workflowを起動するevent](https://docs.github.com/ja/actions/reference/workflows-and-actions/events-that-trigger-workflows)
- [Contextsリファレンス](https://docs.github.com/ja/actions/reference/workflows-and-actions/contexts)
- [安全な利用のリファレンス](https://docs.github.com/ja/actions/reference/security/secure-use)
- [GitHub Actionsの課金](https://docs.github.com/en/billing/concepts/product-billing/github-actions)
- [Runner料金表](https://docs.github.com/en/billing/reference/actions-runner-pricing)
- [Budgetと課金アラート](https://docs.github.com/en/billing/how-tos/set-up-budgets)

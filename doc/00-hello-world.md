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

## 実装前の予想

workflowを書く前に、次の問いへの予想を記録する。

- [ ] workflowファイルはrepository内のどこへ置くか
- [ ] pushされたことをYAMLのどの項目で指定するか
- [ ] jobを動かすOSをどこで指定するか
- [ ] `Hello, world!`を実行するcommandをどこへ書くか
- [ ] repositoryのcheckoutなしで`echo`を実行できるか

## 試したことと結果

壁打ちと実装の進行に合わせて追記する。

## つまずいた点

壁打ちと実装の進行に合わせて追記する。

## ブログへ残したい要点

- CI/CD経験者でも、最初はアプリのbuildから入らず最小workflowの実行単位を確認するとよい
- GitLab CI/CDとの名称の対応だけでなく、実行モデルの違いを実験で確認する

# CI/CD経験者のためのGitHub Actions再入門

GitLab CI/CDの利用経験がある筆者が、GitHub Actionsを基礎から学び直す過程を記録します。
このディレクトリの章別メモを、最終的に同名のブログ記事へ再構成する予定です。

## 想定読者

- CI/CDの基本的な目的は理解している
- GitLab CI/CDなど、GitHub Actions以外のCI/CDを使った経験がある
- GitHub Actionsの用語や設計を対応づけて理解したい

## 章構成

| Chapter | テーマ | Issue | 状態 |
| --- | --- | --- | --- |
| 00 | Hello World | [#3](https://github.com/urchin-hat/study-github-action/issues/3) | 完了 |
| 01 | WorkflowのTrigger | [#4](https://github.com/urchin-hat/study-github-action/issues/4) | 完了 |
| 02 | Contextと変数 | [#5](https://github.com/urchin-hat/study-github-action/issues/5) | 完了 |
| 03 | Jobの依存関係と並列実行 | [#6](https://github.com/urchin-hat/study-github-action/issues/6) | 完了 |
| 04 | サンプルアプリのCI | [#7](https://github.com/urchin-hat/study-github-action/issues/7) | 完了 |
| 05 | CacheとArtifact | [#8](https://github.com/urchin-hat/study-github-action/issues/8) | 完了 |
| 06 | Workflowのセキュリティ | [#9](https://github.com/urchin-hat/study-github-action/issues/9) | 完了 |
| 07 | Environmentを使ったDeploy | [#10](https://github.com/urchin-hat/study-github-action/issues/10) | 完了 |
| 08 | 学習内容の振り返り | [#11](https://github.com/urchin-hat/study-github-action/issues/11) | 完了 |

## 記録方針

各章では、次の観点を残します。

1. その章で理解したいこと
2. GitLab CI/CDの概念との対応
3. 実行前の予想
4. 実際に試した変更と結果
5. つまずいた点と原因
6. ブログへ残したい要点

完成形だけでなく、予想が外れた箇所や失敗から得た理解も記録します。

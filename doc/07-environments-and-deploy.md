# Chapter 07: Environmentを使ったDeploy

対応Issue: [#10 07: Environmentを使ったDeployを設計する](https://github.com/urchin-hat/study-github-action/issues/10)

## この章で理解したいこと

- **GitHub Environment の基本と設計**:
  - `environment: name: ...` によるデプロイ先環境の指定と追跡
  - リポジトリレベルとEnvironmentレベルでのSecrets/Variables分離
  - 承認ルール（Required Reviewers）と保護ルール（Wait timer, Deployment branches）
- **`workflow_dispatch` inputs によるデプロイ先・パラメータの切り替え**:
  - `type: choice` による安全なデプロイ環境選択（staging vs production）
- **`concurrency` による排他制御**:
  - 同じ環境への多重デプロイ防止（`concurrency: group: deploy-${{ inputs.environment }}`）
- **クラウド接続における OIDC（OpenID Connect）の意義**:
  - 静的アクセスキー（永続トークン）を廃止し、一時的な短命トークンを発行する仕組み

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 観点 |
| --- | --- | --- |
| `environment: name: staging` / `url:` | `environment: name: staging` / `url:` | デプロイ先環境の定義と管理 |
| Protected Environments / `when: manual` | Environment Protection Rules (Required reviewers) | デプロイ前の手動承認フロー |
| スコープ付き変数（Environment Variables） | Environment Secrets / Variables | 環境ごとに異なる秘匿値や設定値の注入 |
| `resource_group: production` | `concurrency: group: deploy-${{ ... }}` | 同一環境への同時並行デプロイ防止（排他制御） |
| GitLab CI/CD OIDC (`id_tokens:`) | GitHub Actions OIDC (`permissions: id-token: write`) | クラウドプロバイダとのキーレス認証連携 |

## 最初の方針

1. **デプロイ用ワークフローの作成**:
   - `workflow_dispatch` で `staging` / `production` を選択できるようにする。
2. **Build成果物（Artifact）のデプロイJob連携**:
   - Chapter 05で学んだ `actions/download-artifact` を使い、ビルド済みバイナリを取得してデプロイを模倣する。
3. **GitHub Environment の設定**:
   - `staging` と `production` のEnvironmentを宣言し、UI上でのデプロイ履歴確認を行う。
4. **`concurrency` による同時デプロイの排他制御**:
   - 同じ環境への同時実行がブロックまたはキャンセルされる挙動を確認する。
5. **OIDCの概念理解**:
   - 静的クレデンシャル管理の危険性と、OIDCによるキーレス認証のアーキテクチャを整理する。

## 実装前の予想

- [x] GitLab CI/CDの `resource_group` に相当する「同一環境への多重デプロイ防止」は GitHub Actions ではどう実現するか？:
  - 予想: A（`strategy.max-parallel: 1`）。
  - 実際: **B（`concurrency`）**。`max-parallel` は1つのマトリックスJob内の並列数を絞るだけで、別々のRunやコミット間の排他制御はできない。`concurrency: group: ...` を使うことで、異なるRunやJobをまたいだ同一環境への排他制御・キューイングを実現できる。
- [x] GitHub Environment に設定した Secrets は、その Environment を指定していない Job から参照できるか？:

  - 予想: B（`environment: production` を指定したJobにだけ注入され、指定していないJobからは空文字になる）。
  - 実際: **B（大正解）**。GitLab CI/CDの環境スコープ付き変数と同様、デプロイJob以外からのシークレット漏洩を完全に防ぐ隔離設計になっている。
- [ ] クラウド（AWS / GCP / Azure）へデプロイする際、なぜ永続的なアクセスキー（APIキー）ではなく OIDC を使うべきなのか？

## 壁打ちメモ

### 2026-09-25: Chapter 07開始
- `main` からブランチ `lesson/07-deploy` を作成。

### 2026-09-25: GitHub Environment とスコープ付きSecretsの隔離（予想と壁打ち）
- **Environment Secretsのスコープ制限**:
  - `secrets` をリポジトリ全体に置くと、Lintや単体テストなどのあらゆるJobから読み取れてしまう。
  - `Environment` にシークレット（例: 本番DBのパスワード等）を紐付けると、`environment: <名前>` が明示されたJobの実行時にのみランナーへ安全に渡される。
- **Environment Protection Rules（保護ルール）**:
  - ① **Required reviewers**: 指定したユーザー/チームがWeb UI上で「承認（Review deployments）」を押すまでJobの実行が一時停止し、Secretsも渡されない。
  - ② **Wait timer**: 承認後またはトリガー後、指定分（例: 5分）待機してからデプロイを開始する。
  - ③ **Deployment branches**: 例えば `production` には `main` ブランチからの実行しか許可しない、といった制限が可能。

### 2026-09-25: concurrency による排他制御（GitLabのresource_groupとの対応）
- **GitLab CI/CDの `resource_group`**:
  - 同じ環境へのデプロイが重複して走らないように、先行するJobが終わるまで後続Jobをキュー待ち（待機）させる機能。
- **GitHub Actionsでの実現方法**:
  - `concurrency:` を使用する。
  - 設定例:
    ```yaml
    concurrency:
      group: deploy-${{ inputs.environment }}
      cancel-in-progress: false
    ```
  - `group`: 排他制御を行う識別子。環境名を含めることで、「stagingへのデプロイ」と「productionへのデプロイ」は独立して並行実行させつつ、「同じ環境への同時デプロイ」だけを排他できる。
  - `cancel-in-progress`:
    - `false`（デフォルト）: 先行デプロイが完了するまで後続デプロイは「待機（queued）」する（デプロイに最適）。
    - `true`: 先行するジョブを即座にキャンセルして最新のものだけを実行する（PRのCIやテストに最適）。



## 試したことと結果

### 実験1: `workflow_dispatch` 手動デプロイ、Artifact連携、GitHub Environment 追跡

- PR: [#20](https://github.com/urchin-hat/study-github-action/pull/20)
- Workflow Run: [36082400592](https://github.com/urchin-hat/study-github-action/actions/runs/36082400592)（`environment=staging`）
- 目的:
  - `workflow_dispatch` で環境名（`staging` / `production`）を選択可能にする。
  - `build` ジョブが生成したバイナリアーティファクト（`study-server-binary`）を `deploy` ジョブで `actions/download-artifact` して受け取る。
  - `environment: name: staging, url: https://staging.example.com` を指定し、GitHub Environments のデプロイ履歴に正しく記録されるか確認する。

#### 実行結果
- 各ジョブのステータス: ✅ **All 8 jobs success**
- `Deploy to staging` ジョブの実行ログ:
  ```text
  === Download binary artifact ===
  Redirecting to blob download url: ...
  SHA256 digest of downloaded artifact is 30cb1b21...
  Artifact download completed successfully.

  === Execute Deployment ===
  ==========================================
  🚀 Deploying to Environment: staging
  👤 Deployed by: urchin-hat
  🔖 Commit SHA: bd91e1e2a4b73646fa00e8a16f562f98bd5e58b2
  ==========================================
  -rwxr-xr-x 1 runner runner 7306544 Sep 25 01:32 bin/study-server
  Simulating service deployment...
  ✅ Deployment to staging completed successfully!

  === Complete job ===
  Evaluated environment url: https://staging.example.com
  ```
- GitHub API / UI の確認:
  - `gh api repos/urchin-hat/study-github-action/deployments` で `environment: "staging"` のデプロイレコードが自動作成され、コミットSHAやデプロイ実施者（`creator: urchin-hat`）が記録されたことを確認。

#### 分かったこと
- **成果物の確実な受け渡し**:
  - `build` ジョブでビルドしたバイナリを `upload-artifact` し、`deploy` ジョブで `download-artifact` することで、ビルドとデプロイの明確な関心の分離（ビルド成果物を確実にデプロイする）が実現できた。
- **GitHub Environment によるデプロイ管理**:
  - `environment:` を書くだけで、GitHubが自動的に「Deployments」として認識し、いつ、誰が、どのコミットを、どのURLへデプロイしたかを追跡できる。


## つまずいた点

- **新規ワークフローファイルの `workflow_dispatch` はデフォルトブランチにマージされるまで使えない**:
  - 新規に `.github/workflows/deploy.yml` を作成してブランチへプッシュしても、GitHub Actionsの仕様上 `HTTP 404: workflow deploy.yml not found on the default branch` となり、手動実行することができない。
  - そのため、学習段階やPR段階で手動デプロイを検証するには、すでに `main` に存在する `.github/workflows/ci.yml` に `workflow_dispatch` の `inputs` と `deploy` ジョブを追加・統合して実験を進めるのが確実である。

## ブログへ残したい要点



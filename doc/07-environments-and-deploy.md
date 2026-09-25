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
- [x] クラウド（AWS / GCP / Azure）へデプロイする際、なぜ永続的なアクセスキー（APIキー）ではなく OIDC を使うべきなのか？:
  - 予想: B（長期鍵をGitHub Secretsに保管する必要がなくなり、漏洩リスクや定期ローテーション運用がゼロになるから）。
  - 実際: **B（大正解）**。GitHub Actions（OIDCプロバイダ）が発行する短命な署名付きJWTトークンをクラウド（AWS/GCP）に提示し、一時的な認証情報を取得する「完全キーレス（Keyless）認証」が実現できるため。

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

### 2026-09-25: OIDC（OpenID Connect）によるキーレス認証のアーキテクチャ
- **静的アクセスキーの課題**:
  - AWS IAMユーザーのアクセスキーやGCPサービスアカウントキーをGitHub Secretsに置くと、万が一の漏洩リスクや90日ごとの定期ローテーション運用が発生する。
- **OIDCの仕組み**:
  - GitHub ActionsランナーがGitHubのOIDCプロバイダから短命な署名付きJWTトークンを取得。
  - そのトークンをクラウド側（AWS AssumeRoleWithWebIdentity / GCP Workload Identity Federation）に提示し、一時的なアクセスキーを払い出してもらう。
  - クラウド側のIAMロール側で「リポジトリ名」「ブランチ」「Environment」などのクレーム（条件）を検証できるため、安全性が極めて高い。
  - ワークフロー側には `permissions: id-token: write` を付与する。

### 2026-09-25: チーム開発におけるCI/CD設計の勘所と高速化
- **チーム運用の勘所**:
  - ① **フィードバックの多段階（グラデーション）設計**: PR時は高速チェック（3〜5分以内）、マージ時や夜間に重いE2E・セキュリティスキャン。
  - ② **ローカルとCIのコマンド共通化**: MakefileやTaskfileで `make test` や `make lint` を揃え、「CIでしか動かない」を防ぐ。
  - ③ **権限と環境の分離**: PRには最小権限（readのみ）、デプロイはEnvironment + OIDC。
  - ④ **共通化（Reusable Workflows）**: 組織横断でのYAMLの重複を排除。
- **テスト肥大化・パイプライン遅延の処方箋**:
  - ① **差分実行（Path filtering）**: `paths` や `dorny/paths-filter` で変更のあった箇所のみテスト。
  - ② **キャッシュ多層化**: 依存関係（`go mod`）だけでなくビルドキャッシュやDockerレイヤーキャッシュも保持。
  - ③ **テストのシャーディング**: `strategy.matrix` でテストスイートを並列分割実行。
  - ④ **不要な過去Runのキャンセル**: PRのCIには `concurrency: cancel-in-progress: true`。
  - ⑤ **テストピラミッドの見直し**: 重いE2Eテストを高速なUnitテストへ寄せる。

### 2026-09-25: Google SREの行動原理に基づくアプローチ
- **CI/CDパイプラインは開発組織にとっての本番サービス**:
  - ① **SLI/SLO とエラーバジェット**:
    - 「CIが遅い」を定量化（例: PR CIのP95完了時間を5分以内）。バジェット枯渇時は新機能開発を止め、全員でCI高速化・健全化に取り組む合意。
  - ② **トイル（Toil）の撲滅**:
    - 「落ちたら Re-run を押す」「手動でテストをスキップする」などの手作業を許さず、根本原因をエンジニアリングで解決。
  - ③ **Flakyテストへのゼロトレランス（自動隔離 - Quarantine）**:
    - 不安定なテストは開発者のアラート疲労を生む最悪の要因。自動検知して隔離レーンへ送り、担当者にチケットを起票して修正されるまで本流から外す。
  - ④ **オブザーバビリティ（計測とデータ主導）**:
    - 各ステップの所要時間やキュー待ち時間を可視化し、クリティカルパスを特定して投資する。
  - ⑤ **Paved Path（舗装された道）の提供**:
    - SREは門番ではなく、乗るだけで最速・セキュアになる標準ワークフロー基盤を提供するイネーブラとなる。

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

### 実験2: `concurrency` による同一環境への排他制御（キュー待ち）の実証

- PR: [#20](https://github.com/urchin-hat/study-github-action/pull/20)
- 実行Run:
  - Run 1: [36082748285](https://github.com/urchin-hat/study-github-action/actions/runs/36082748285)
  - Run 2: [36082750821](https://github.com/urchin-hat/study-github-action/actions/runs/36082750821)
- 目的:
  - `concurrency: group: deploy-${{ inputs.environment }}, cancel-in-progress: false` を設定。
  - `staging` への手動デプロイをほぼ同時に2回連続でトリガーし、`deploy` ジョブが重複並行実行されず、直列化（キュー待ち）されるかを確認する。

#### 実行結果
- 各Runの `Deploy to staging` ジョブの実行タイムスタンプ比較:
  - **Run 1**: `startedAt: 01:37:16Z` 〜 `completedAt: 01:37:39Z`（約23秒実行）
  - **Run 2**: `startedAt: 01:37:45Z` 〜 `completedAt: 01:38:11Z`（約26秒実行）
- 観測された挙動:
  - 先行ジョブ（LintやTest）は並列に実行されたが、`Deploy to staging` に入った瞬間、同じグループ `deploy-staging` であるため **Run 2 は Run 1 の完了を待機（queued）** した。
  - Run 1 が `01:37:39Z` に成功した直後の `01:37:45Z` から Run 2 のデプロイが自動的に開始され、正常終了した。

#### 分かったこと
- **GitLab CI/CDの `resource_group` と完全同等の排他制御**:
  - `concurrency: group: ...` と `cancel-in-progress: false` を組み合わせることで、同一環境への多重デプロイや順序逆転の事故を確実に防ぐことができる。
  - `group` 名に環境変数や入力値（`inputs.environment`）を含めることで、環境単位（staging同士、production同士）での独立した排他制御が可能になる。

## つまずいた点

- **新規ワークフローファイルの `workflow_dispatch` はデフォルトブランチにマージされるまで使えない**:
  - 新規に `.github/workflows/deploy.yml` を作成してブランチへプッシュしても、GitHub Actionsの仕様上 `HTTP 404: workflow deploy.yml not found on the default branch` となり、手動実行することができない。
  - そのため、学習段階やPR段階で手動デプロイを検証するには、すでに `main` に存在する `.github/workflows/ci.yml` に `workflow_dispatch` の `inputs` と `deploy` ジョブを追加・統合して実験を進めるのが確実である。

## ブログへ残したい要点

1. **GitHub Environment による安全なデプロイ設計**:
   - `environment: name: staging` と書くだけで、GitHub Deploymentsと自動連携され、デプロイ履歴やURLの追跡が可能になる。
   - Environment単位でSecretsや保護ルール（Required Reviewers承認、待機時間、実行ブランチ制限）を隔離設定でき、本番クレデンシャルの漏洩リスクを最小化できる。
2. **GitLab CI/CD `resource_group` vs GitHub Actions `concurrency`**:
   - 同一環境への多重デプロイを防ぐ排他制御は、`concurrency: group: ...` で実現できる。
   - デプロイジョブには `cancel-in-progress: false` を指定することで、先行デプロイをキャンセルせず安全にキュー待ち（直列化）させられる。
3. **OIDC（OpenID Connect）による完全キーレス認証**:
   - 静的アクセスキーをGitHub Secretsに置く時代は終わり、`permissions: id-token: write` による短命JWTトークン連携がクラウドデプロイのデファクトスタンダード。
4. **持続可能なCI/CD運用とSRE的アプローチ**:
   - CI/CDパイプラインも本番サービスとして捉え、SLI/SLO（P95 ≤ 5分）とエラーバジェットで速度と品質のバランスを取る。
   - トイルの撲滅、Flakyテストの自動隔離（Quarantine）、テストシャーディング（並列分割）、そして安全・高速なPaved Path（舗装された道）の提供が開発組織の生産性を最大化する。



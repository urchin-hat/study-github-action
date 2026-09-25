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

- [ ] GitLab CI/CDの `resource_group` に相当する「同一環境への多重デプロイ防止」は GitHub Actions ではどう実現するか？
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


## 試したことと結果

## つまずいた点

## ブログへ残したい要点


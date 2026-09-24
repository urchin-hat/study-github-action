# Chapter 06: Workflowのセキュリティ

対応Issue: [#9 06: Workflowのセキュリティを学ぶ](https://github.com/urchin-hat/study-github-action/issues/9)

## この章で理解したいこと

- **`permissions` による `GITHUB_TOKEN` の最小権限化**:
  - デフォルト権限（Read/Write）のリスクと、`permissions: read-all` / `permissions: contents: read` 等による絞り込み
  - WorkflowレベルとJobレベルでのスコープ制御
- **`vars` と `secrets` の使い分けと保護**:
  - 非機密設定値（`vars`）と機密情報（`secrets`）の適切な分離
  - ログ出力時の自動マスキング（`***`）の挙動と限界
- **ForkリポジトリからのPRとセキュリティ境界**:
  - ForkからのPRで `secrets` が渡らない理由と安全設計
  - `pull_request` vs `pull_request_target` の違いと重大な脆弱性パターン（PWN request）
- **未信頼の入力とスクリプトインジェクション（Script Injection）**:
  - `${{ github.event.issue.title }}` や `${{ github.head_ref }}` を `run:` に直接埋め込む危険性
  - 環境変数（`env:`）を経由した安全な渡し方
- **サードパーティActionのサプライチェーンセキュリティ**:
  - Actionの発行元（公式 / Verified Creator / 個人）とバージョン指定（tag vs commit SHA）

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 観点 |
| --- | --- | --- |
| `CI_JOB_TOKEN` (パーミッション制限機能あり) | `GITHUB_TOKEN` + `permissions:` | ジョブ実行時の組み込みトークンと権限管理 |
| CI/CD Variables (Masked / Protected) | GitHub Secrets / Repository Variables (`vars`) | 機密情報と非機密設定値の管理 |
| Protected Branches / Environments | `pull_request` の制限 / Environment Protection Rules | 実行環境やブランチに応じたシークレットアクセス制御 |
| スクリプト内での変数展開リスク | `${{ ... }}` の直接展開（インジェクション脆弱性） | 外部入力の展開リスク |
| `include:` での外部CI定義取り込み | サードパーティAction（`uses: ...`） | 外部コードの取り込みとサプライチェーンリスク |

## 最初の方針

1. **`permissions` の実験**:
   - 現在のワークフローに明示的な `permissions` を設定し、最小権限（`contents: read`）にする。
2. **スクリプトインジェクションの実験**:
   - PRのタイトルやコミットメッセージなどを `run: echo "${{ ... }}"` で直接展開した場合と、`env:` 経由で渡した場合の安全性の違いを確認する。
3. **Secrets / Vars のマスキング確認**:
   - `secrets` を渡した際のログマスキング挙動を確認し、露出しない安全な扱い方を検証する。

## 実装前の予想

- [ ] `permissions` をトップレベルで宣言しない場合、`GITHUB_TOKEN` にはどのような権限が与えられているか？
- [ ] `${{ github.event.pull_request.title }}` を `run: echo "${{ ... }}"` に直接書くと、どのような攻撃（インジェクション）が可能になるか？
- [ ] `pull_request` と `pull_request_target` で `secrets` や `GITHUB_TOKEN` の権限はどう違うか？

## 壁打ちメモ

### 2026-09-25: Chapter 06開始
- `main` からブランチ `lesson/06-security` を作成。

## 試したことと結果

## つまずいた点

## ブログへ残したい要点


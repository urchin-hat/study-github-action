# Chapter 08: 学習内容の振り返りとトラブルシューティング

対応Issue: [#11 08: 学習内容を振り返る](https://github.com/urchin-hat/study-github-action/issues/11)

## この章で理解したいこと

- **CI/CDパイプライン全体の実行グラフと構造の体系的理解**:
  - `workflow`, `job`, `step`, `action`, `runner` の階層構造と役割
  - Trigger, Context, `needs`, `cache`, `artifact`, `permissions`, `concurrency`, `environment` の使い分け
- **トラブルシューティングと障害調査の実践**:
  - YAML構文エラーの検知と失敗箇所の特定
  - Shellコマンド失敗時のログ解析とエラーハンドリング
  - Job依存関係（`needs`）によるDownstreamジョブのスキップ挙動
  - Cache miss と Artifact 不足（未生成・期限切れ）の本質的な違い
  - 全体再実行（Re-run all jobs）と失敗ジョブのみの再実行（Re-run failed jobs）の挙動
- **運用・コスト管理と今後の発展テーマ**:
  - GitHub Actionsの利用時間（Minutes）とArtifactストレージ使用量の確認
  - Self-hosted Runner、Composite Actions、Reusable Workflows などの発展的トピック

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 観点 |
| --- | --- | --- |
| `.gitlab-ci.yml` 全体 | `.github/workflows/*.yml` | ワークフロー定義（GitLabは1ファイル統合、GitHubは複数ファイル分割可能） |
| `stages:` / `stage:` / `needs:` | `jobs.<id>.needs` | ジョブの依存関係と実行DAG（有向非巡回グラフ）の構築 |
| `script:` / `before_script:` | `steps:` (`run:` / `uses:`) | コマンド実行ステップと再利用可能なActionの組み合わせ |
| GitLab Runner (Tags) | `runs-on: ubuntu-latest` | 実行マシンの選択とプロビジョニング |
| `Retry failed jobs` | `Re-run failed jobs` | 失敗したジョブのみの選択的再実行 |
| Usage Quotas (Compute minutes / Storage) | Billing & plans / Actions storage | 実行時間とアーティファクト保存容量の監視と最適化 |

## 最初の方針

1. **ワークフロー全体構造の可視化と総括**:
   - `ci.yml` のDAG（Lint/Test -> Build -> Verify -> Deploy）の依存グラフを整理する。
2. **トラブルシューティング実験1: YAML構文エラー**:
   - インデントミスや無効なキーを導入し、GitHub Actionsがどの段階（パース段階）でどのようにエラーを表示するか確認する。
3. **トラブルシューティング実験2: 意図的なコマンド失敗と依存Jobのスキップ**:
   - 単体テスト等を意図的に失敗させ、後続の `build`, `verify`, `deploy` がどうスキップされるかを観察する。
4. **トラブルシューティング実験3: 失敗ジョブの再実行（Re-run failed jobs）**:
   - 失敗したジョブのみを再実行し、成功済みのジョブがスキップされて時間を節約できるかを確認する。
5. **Cache Miss vs Artifact 不足の違いの再確認**:
   - 設計ミスや期限切れ時の挙動（Cache missはリカバリ可能、Artifact欠損はパイプライン停止）を明確にする。
6. **Actionsリソース使用量の確認と総まとめ**:
   - GitHub UI / CLIでストレージや利用時間を確認し、全体の学びを総括する。

## 実装前の予想

- [x] `needs: [build]` を持つ `verify` ジョブは、`build` が失敗した場合どのようなステータスになるか？:
  - 予想: B（スキップされる）。
  - 実際: **B（大正解）**。デフォルトで暗黙の `if: success()` が適用されるため、依存ジョブの失敗時は即座に `Skipped` になる。
- [x] キャッシュ（Cache）が存在しない場合と、アーティファクト（Artifact）が存在しない場合、パイプラインの成否はどう異なるか？:
  - 予想: B（Cacheは再生成可能でパイプライン継続、Artifactは後続処理で必須のためパイプライン停止）。
  - 実際: **B（大正解）**。Cache miss はビルド時間が延びるだけで成功するが、Artifact欠損は後続ジョブがファイルを見つけられず即エラー終了する。
- [x] 失敗ジョブの再実行（Re-run failed jobs）をした場合、成功済みの先行ジョブのキャッシュやアーティファクトは後続ジョブで再利用できるか？:
  - 予想: 再利用できる。
  - 実際: **そのとおり**。成功済みの先行ジョブはスキップされ、先行ジョブが生成・アップロードしたArtifactを再実行された後続ジョブがそのままダウンロードして処理できる。

## 壁打ちメモ

### 2026-09-25: Chapter 08開始
- `main` からブランチ `lesson/08-retrospective` を作成。

### 2026-09-25: ワークフロー全体のDAG構造
- 完成した `.github/workflows/ci.yml` のジョブ依存グラフ:
  ```mermaid
  flowchart TD
      Trigger(["Trigger: push / pull_request / workflow_dispatch"]) --> Lint["lint (Lint & Format)"]
      Trigger --> Test["test (Unit Test: Go 1.21, 1.22, 1.23)"]
      Trigger --> Security["security_injection_test"]

      Lint --> Build["build (Build Binary & Cache & Upload Artifact)"]
      Test --> Build

      Build --> Verify["verify (Download Artifact & Health Check)"]
      Build -.->|if: workflow_dispatch| Deploy["deploy (Environment: staging/prod, Concurrency)"]

      style Trigger fill:#f9f,stroke:#333,stroke-width:2px
      style Build fill:#bbf,stroke:#333
      style Deploy fill:#bfb,stroke:#333
  ```

### 2026-09-25: ローカル検証・構文確認のお作法（プッシュ前に防ぐ）
- 「プッシュしてGitHub画面を見てタイポに気づく」というトイルを撲滅するためのツール:
  - ① **`actionlint`（デファクトスタンダード）**:
    - Go製の超高速静的解析ツール。
    - YAML構文、`${{ ... }}` 式構文の型チェック、未定義コンテキストの参照検出。
    - `run:` 内のシェルスクリプトを **ShellCheck** と自動連携して文法・脆弱性（インジェクションリスク等）を検査。
    - `brew install actionlint` で導入し、`pre-commit` フックや手元CLIで回すのが実務のお作法。
  - ② **`act` (nektos/act)**:
    - ローカルのDocker環境を使って、GitHub Actionsの実行環境を手元でエミュレート。
    - `act -l`: 定義されているジョブ一覧の表示。
    - `act -n`: 実際にコンテナを起動せずに実行計画をシミュレーション（Dry-run）。
    - `act -j <job-id>`: 特定のジョブだけを手元コンテナで実行してデバッグ。
  - ③ **GitHub Actions 公式拡張（VS Code）**:
    - エディタ上でリアルタイムにスキーマ検証・自動補完・Secrets名補完を提供。
  - **GitLab CI/CDとの対比**:
    - GitLabはWeb UI上の「CI Lint」ツールやAPI (`/ci/lint`) が中心だったが、GitHub Actionsは `actionlint` によるローカル完結の検証が主流。

### 2026-09-25: ワークフローを複数に分割する設計判断（可読性とアンチパターン）
- **GitLab CI/CDとの発想の違い**:
  - GitLab CI/CD: 1つの `.gitlab-ci.yml` にパイプライン全体を集約し、StageやRulesで分岐する文化。
  - GitHub Actions: `.github/workflows/*.yml` 配下に、**関心事・トリガー・権限境界ごとに独立したWorkflowファイルへ分割する** 文化。
- **積極的に分割すべき5大パターン**:
  - ① **ライフサイクル / トリガー別**:
    - PRチェック（`on: pull_request`）と本番デプロイ（`on: push: tags`）は分ける（1ファイルにまとめると `if:` だらけになる）。
  - ② **実行頻度・時間軸別 (Fast Feedback vs Heavy Test)**:
    - 毎回のPRで走る軽量CI（3分）と、深夜に回す重いE2Eテスト・負荷テスト（`on: schedule` / 30分）は分ける。
  - ③ **セキュリティ境界（権限）の分離**:
    - 一般PR向けの Read-only ワークフローと、パッケージ公開やデプロイ用の特権ワークフロー（`permissions: id-token: write` 等）を物理ファイルとして分離。
  - ④ **モノレポ（Monorepo）でのコンポーネント別**:
    - `frontend-ci.yml`（`paths: ['frontend/**']`）と `backend-ci.yml`（`paths: ['backend/**']`）。
  - ⑤ **GitHub Ops / 自動化**:
    - `labeler.yml`（PRラベル自動付与）、`stale.yml`（休眠Issue自動クローズ）、`dependabot-auto-merge.yml`。
- **分割しすぎのアンチパターン（やりすぎの弊害）**:
  - ❌ **「1つのPRで動く一連の依存フロー」をファイルごとに細切れにする**:
    - 例: `pr-lint.yml`, `pr-test.yml`, `pr-build.yml` とバラバラにする。
    - **弊害1**: `needs:` は「同一ファイル内のJob間」でしか使えないため、依存制御が極めて困難になる。
    - **弊害2**: PR画面に大量のWorkflow Runが乱立し、Checks画面がノイズまみれになる。
    - **弊害3**: 成果物（Artifact）の受け渡しが難しくなる。
- **💡 黄金の判断基準**:
  - **「1つのイベントで始まり、`needs` で順序制御したり、Artifactを受け渡し合いたい一連のフロー」** は、1つのWorkflowファイルにまとめ、内部で複数のJob（DAG）として並べる。
  - **「それ以外の、イベント・権限・実行タイミングが異なるもの」** は、積極的に別ファイルへ分割する。

### 2026-09-25: 厳選4問の総復習（GitLab CI/CDとのギャップ・つまずきポイント）
- **Q1: Cache vs Artifact の本質**:
  - `Cache`: 高速化目的、消えてもOK（Miss時はゼロから再生成して継続）。
  - `Artifact`: 成果物受け渡し目的、消えたら困る（欠損時はパイプライン停止）。GitLabと異なり明示的な `actions/download-artifact` が必須。
- **Q2: GITHUB_TOKEN のデフォルト権限**:
  - 歴史的経緯から、リポジトリやOrganizationの設定次第で「Read & write（読み書き全開放）」になっていることがある。
  - トップレベルで `permissions: contents: read` を宣言して最小権限化するのが絶対必須。
- **Q3: 同一環境への多重デプロイ防止（排他制御）**:
  - `strategy.max-parallel` は単一Run内のMatrix並列数制御のみ。
  - 異なるコミットや別々のRunをまたいだ排他制御（キュー待ち）には `concurrency: group: ...` が必要（GitLabの `resource_group` 相当）。
- **Q4: Trigger における `branches: [main]` の評価対象の違い**:
  - `push.branches: [main]`: pushされたコミット自身のブランチを判定（作業ブランチへのpushでは動かない）。
  - `pull_request.branches: [main]`: PRの **取り込み先（base branch）** を判定（作業ブランチからのPRでも動く）。

### 2026-09-25: 運用面・コストの勘所（課金爆発を防ぐ）
- **① RunnerのOS別コスト倍率（macOSの罠）**:
  - Linux (Ubuntu): 1倍（標準・最安）
  - Windows: 2倍
  - macOS: **10倍**（大型M1/M2ランナーはそれ以上）
  - 対策: Lintや単体テストなどの非依存処理はLinuxランナーで動かし、macOSランナーは最後のIPA/APKビルドのみに限定する。
- **② タイムアウト未指定による「6時間放置」の課金事故**:
  - デフォルトの `timeout-minutes` は **360分（6時間）**。stdin待ちやハングで枠が溶けるのを防ぐため、全Jobに **`timeout-minutes: 5` 〜 `15`** を明示する。
- **③ Artifactの保存期間（retention-days）によるストレージ課金**:
  - デフォルト保存期間は **90日間**。毎回数百MBを保存するとアカウントのストレージ枠を食い潰す。PR検証用は **`retention-days: 1` 〜 `3`** に短縮する。
- **④ PRプッシュ連打によるランナー枠占有**:
  - PRには **`concurrency: cancel-in-progress: true`** を設定し、最新コミット以外の古い実行を即座に自動キャンセルさせる。

### 2026-09-25: 現場で引っかかりやすい落とし穴（実務あるある）
- **① 外部Actionのタグ指定リスク**:
  - タグ（`@v4`）は可変。コミットSHAで固定（`uses: actions/checkout@b4ffde...`）し、Dependabotで自動更新するのが安全。
- **② `pull_request` vs `pull_request_target` のセキュリティ地雷**:
  - `pull_request`: ForkからのPRではSecrets空、`GITHUB_TOKEN` はread-only（安全）。
  - `pull_request_target`: 親リポジトリの全Secretsと書き込み権限で動作する。Fork元のコードをチェックアウトしてスクリプト実行するとSecretsが漏洩する（PwnRequest攻撃）。
- **③ レートリミット（Docker Hub / GitHub API）**:
  - ランナーのIP共有による制限。自社レジストリ（ghcr.io）移行やキャッシュ活用で回避。
- **④ cron定期実行の「60日休眠ルール」**:
  - リポジトリに60日間コミットがないと、`schedule` ワークフローは自動で無効化される。

### 2026-09-25: ログ調査・RAWログ確認のTips（GitLabとの比較）
- GitLab CI/CDのプレーンテキストログに慣れていると、GitHub ActionsのアコーディオンUIは一覧検索しづらい。
- **解決策1: Web UI の RAW ログ表示**:
  - ジョブ画面右上の **歯車アイコン（⚙️） ➜ `View raw logs`** を開くと、別タブで完全プレーンテキストが開く（`Ctrl+F` で全体一括検索可能）。
  - `Download log archive` で全ジョブの生ログを一括ZIP取得可能。
- **解決策2: GitHub CLI (`gh`) によるターミナル閲覧（最速）**:
  - `gh run view <run-id> --log-failed`: 失敗したステップのログのみをピンポイント抽出。
  - `gh run view <run-id> --log | grep -i "error"`: ターミナルから直接検索。

### 2026-09-25: CI/CDのSLI/SLO数値の取得方法（オブザーバビリティ基盤）
- **レベル1: GitHub公式「Insights」タブ（手軽な可視化）**:
  - Web UIの「Insights」➜「Actions」で、ワークフロー別の成功率（Success rate）や平均実行時間（Run duration）の推移をグラフ確認可能。
- **レベル2: GitHub REST API / CLI からの自前集計（正確な算出）**:
  - `gh api repos/{owner}/{repo}/actions/runs` で各Runのタイムスタンプを取得し、以下を算出可能：
    - **キュー待ち時間**: `run_started_at - created_at`
    - **実行時間（Duration）**: `updated_at - run_started_at`
    - リストをソートして **P95実行時間** や **成功率（SLO達成率）** を算出。
- **レベル3: SREの本番運用！「CIオブザーバビリティ基盤」への流し込み（業界標準）**:
  - ① **Datadog CI Visibility / New Relic CI/CD Visibility**:
    - Webhookや公式Action（`newrelic/github-actions-telemetry-action`）を登録するだけで、P95完了時間、キュー時間、そして最大の敵である **「Flakyテスト（不安定テスト）の自動検知ランキング」** を自動提供。
    - New RelicではNRQL（`SELECT percentile(duration, 95) FROM GitHubActionsWorkflowJob`）を使って本番サーバーのAPMと同じ画面でCIのSLOを一元監視可能。
  - ② **GitHub Webhooks (`workflow_run`, `workflow_job`) + 自社監視**:
    - イベントをLambda経由でPrometheus / BigQuery / CloudWatchへ投入し、Grafanaで本番サービスと同じダッシュボード＆Slackアラートを構築。

### 2026-09-25: 開発指標として取得できる重要メトリクス（DORAと健康診断）
- **① 世界標準！DORAメトリクス（Four Keys）**:
  - **デプロイ頻度 (Deployment Frequency)**: `environment: production` の完了数（1日○回/週○回）。
  - **変更のリードタイム (Lead Time for Changes)**: コミット作成から本番デプロイ完了までの時間。
  - **変更障害率 (Change Failure Rate)**: 本番デプロイ後にロールバックやHotfixが必要になった割合。
  - **サービス復旧時間 (MTTR / Time to Restore Service)**: 本番障害発生から修正デプロイ完了までの復旧時間。
- **② CI / テストパイプラインの健康度**:
  - **キュー待ち時間 (Queue Wait Time)**: PRを出してからランナーが起動するまでの待ち時間（ランナー枠不足の検知）。
  - **Flakiness率 (Flaky Test Rate)**: リトライで通った不安定テストの割合（テスト信頼性の指標）。
  - **キャッシュヒット率 (Cache Hit Ratio)**: ビルドキャッシュが正常に効いているかの監視。
- **③ PR / レビュープロセスの健全度**:
  - **PRサイクルタイム (Time to Merge)**: PR作成からマージまでの総時間（プロセスの詰まり検知）。
  - **初回レビューまでの時間 (Time to First Review)**: PR作成から最初のレビューがつくまでの時間。
  - **PRサイズ (Lines of Code)**: 差分行数（300行以内を推奨、巨大PRはバグと遅延の主因）。

## 試したことと結果

### 1. リソース使用量（Cache & Artifact Storage）の確認
- GitHub CLI（API）を用いて、リポジトリ内のストレージ使用状況を確認した。
  ```bash
  # Cache 使用量の確認
  gh api repos/urchin-hat/study-github-action/actions/cache/usage
  ```
  - **結果**:
    - `active_caches_count`: 2
    - `active_caches_size_in_bytes`: 31,772,538 bytes（約 31.7 MB）
    - ※リポジトリ全体で最大 10 GB まで無料枠で利用可能。
  ```bash
  # Artifact 一覧と容量の確認
  gh api repos/urchin-hat/study-github-action/actions/artifacts
  ```
  - **結果**:
    - ビルドした各バイナリ（`study-server-binary`）が約 4.18 MB ずつ保存されている。
    - ワークフロー側で `retention-days: 1` を指定したため、24時間後に自動的に `expired: true` となり容量を圧迫しない設計になっていることを確認。

## つまずいた点（全Chapter総まとめ）

1. **Chapter 01 (Triggers)**:
   - `github.ref` はPR実行時、作業ブランチではなく `refs/pull/<PR番号>/merge` という合成コミットを参照する。
   - 新規作成したワークフローの `workflow_dispatch` は、デフォルトブランチ（main）にマージされるまでWeb UIやAPIから実行できない。
2. **Chapter 02 (Context & Variables)**:
   - `${{ ... }}` をecho文の文字列としてそのまま書くと、パーサーが式として解釈して構文エラーになる（`${{ '${{ ... }}' }}` とエスケープが必要）。
   - 式言語の文字列リテラルはシングルクォート `'...'` のみ。
3. **Chapter 03 (Job Design)**:
   - GitLab CI/CDの `stages` は存在せず、すべてのJobはデフォルトで完全並列実行される。順序制御には `needs:` が必須。
   - Jobごとに新しい独立した仮想マシンが割り当てられるため、ファイルやメモリは一切引き継がれない。
4. **Chapter 05 (Cache & Artifact)**:
   - GitLab CI/CDのように成果物が自動で後続ジョブに渡らない。必ず `actions/download-artifact` が必要。
5. **Chapter 06 (Security)**:
   - Context（PRタイトル等）を `run:` 内に直接埋め込むと、スクリプトインジェクションが成立する。必ず `env:` を介して渡す。
6. **Chapter 07 (Environments)**:
   - 並列数制限の `max-parallel` では多重デプロイは防げない。`concurrency` を使う必要がある。

## ブログへ残したい要点

### GitLab CI/CD経験者のためのGitHub Actions完全対比マップ

| 観点 | GitLab CI/CD | GitHub Actions | 移行時の重要ポイント |
| :--- | :--- | :--- | :--- |
| **構造** | 単一の `.gitlab-ci.yml` | `.github/workflows/*.yml`（複数ファイル） | 責務ごとにファイル分割可能 |
| **起動条件** | `workflow:rules` / `rules:` | `on:` / `if:` | `on` はRun自体の生成、`if` はJob/Stepの実行可否 |
| **依存関係** | `stages:` + `needs:` | `needs:` のみ（DAGモデル） | 指定しないと全Jobが完全並列で走る |
| **環境変数** | すべてシェル環境変数として注入 | Context (`${{ }}`) と環境変数 (`$VAR`) が分離 | 外部入力Contextは必ず `env:` を通す（インジェクション対策） |
| **成果物** | `artifacts:`（次ステージへ自動展開） | `upload-artifact` / `download-artifact` | 明示的にダウンロードステップを書く必要がある |
| **キャッシュ** | `cache:`（キーに基づく復元） | `actions/cache` | Missしてもパイプラインは継続する（高速化目的） |
| **トークン権限** | ジョブ単位のRole/Permission | `permissions:` | トップレベルで `permissions: contents: read` を徹底 |
| **デプロイ排他** | `resource_group:` | `concurrency: group: ...` | `cancel-in-progress: false` で安全に直列化 |
| **クラウド認証** | `id_tokens:` (OIDC) | `permissions: id-token: write` | 完全キーレス認証（短命JWTトークン）が標準 |
| **ローカル検証** | Web UI の CI Lint | `actionlint` / `act` | プッシュ前にローカルで静的解析・Dry-runするのがお作法 |
| **SRE運用** | パイプライン時間のモニタリング | Actions API / SLI/SLO 管理 | CIも本番サービス。P95 ≤ 5分、Flakyテストは自動隔離（Quarantine） |

### 開発・運用で迷わないための2大ベストプラクティス

1. **プッシュ前の構文チェック（トイル撲滅のお作法）**:
   - プッシュ後にGitHub上で構文エラーに気づくのは時間と実行枠の浪費。
   - `actionlint`（スキーマ・式構文・ShellCheck）を手元やpre-commitで自動実行し、重いジョブは `act` でDockerローカル実行して事前に潰す。
2. **ワークフロー分割の黄金ルール（可読性と依存関係の両立）**:
   - **分割すべきもの**: トリガーが違う（PR vs Tag）、権限が違う（Read vs Write/OIDC）、実行時間帯が違う（PR軽量CI vs 深夜E2E）。
   - **1ファイルにまとめるべきもの**: 1つのイベントで始まり、`needs:` による順序制御や Artifact の受け渡しが発生する一連の処理（細切れにしすぎると `needs` が使えずRunが乱立する）。



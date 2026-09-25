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

- [ ] `needs: [build]` を持つ `verify` ジョブは、`build` が失敗した場合どのようなステータスになるか？
- [ ] キャッシュ（Cache）が存在しない場合と、アーティファクト（Artifact）が存在しない場合、パイプラインの成否はどう異なるか？
- [ ] 失敗ジョブの再実行（Re-run failed jobs）をした場合、成功済みの先行ジョブのキャッシュやアーティファクトは後続ジョブで再利用できるか？

## 壁打ちメモ

### 2026-09-25: Chapter 08開始
- `main` からブランチ `lesson/08-retrospective` を作成。

## 試したことと結果

（実験後に記録）

## つまずいた点

（実験後に記録）

## ブログへ残したい要点

（実験後に記録）

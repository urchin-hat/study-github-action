# Chapter 04: サンプルアプリのCIを作る（Go）

対応Issue: [#7 04: サンプルアプリのCIを作る](https://github.com/urchin-hat/study-github-action/issues/7)

## この章で理解したいこと

- 公式Action（`actions/checkout`, `actions/setup-go`）を使った実コードの取得と言語ランタイムのセットアップ
- GitLab CI/CDのコンテナイメージ指定（`image:`）とGitHub Actionsの「ホストRunner + Setup Action」のアプローチの違い
- `lint`, `test`, `build` の実務的なCI Job分離とDAG依存関係（`needs`）
- `strategy.matrix` による複数言語バージョンの並列テスト
- `fail-fast` の挙動（デフォルトのキャンセル動作と制御）

## GitLab CI/CDとの対応

| GitLab CI/CD | GitHub Actions | 観点 |
| --- | --- | --- |
| `image: golang:1.22` | `runs-on: ubuntu-latest` + `actions/setup-go@v5` | 実行環境・言語ランタイムの準備 |
| リポジトリは自動Clone | `actions/checkout@v4` が明示的に必要 | ソースコードの取得 |
| `parallel: matrix` | `strategy: matrix` | 複数バージョン等の組み合わせ並列実行 |
| `parallel: matrix` は個別完走 | `strategy.fail-fast: true`（既定値） | matrixの1つがコケた時の他ジョブのキャンセル |
| `cache: paths: [/go/pkg/mod]` | `actions/setup-go` 内蔵キャッシュ | 言語依存関係のキャッシュ |

GitLab CI/CDではJobごとにDockerコンテナイメージ（`image: golang:1.22`）を直接指定してその中でコマンドを実行するのが標準的ですが、GitHub Actionsでは **VM型Runner（`ubuntu-latest`）の上で公式の `setup-*` Actionを使って言語環境を構成する** のが標準的なアプローチです。

## 最初の方針

1. サンプルアプリ（Go言語の最小Webサーバー＋ユニットテスト）を作成する
2. `actions/checkout` と `actions/setup-go` を使って、リポジトリのコード取得とGo実行環境のセットアップを行う
3. `lint`（または `go vet` / `test`）、`build` のJobを分離して組み立てる
4. `strategy.matrix` を導入し、Goの複数バージョン（例: 1.22, 1.23）で並列テストを実行する
5. `fail-fast` の挙動（デフォルトで他がどうなるか）を検証する

## 実装前の予想

- [x] リポジトリのコードは、GitLab CI/CDのようにJob開始時に自動でRunner上にCloneされているか？それとも明示的な手順が必要か？: 予想は明示的な手順が必要（B）。GitLabと異なり `actions/checkout` を呼ばないとRunnerのワークスペースは空っぽになる。
- [x] `ubuntu-latest` RunnerにはGoが最初から入っているか？なぜ `setup-go` を使うのか？: 予想は「デフォルトでGoは入っている（A）」。実際もデフォルトでGoやNode.js等はプリインストールされているが、実務ではバージョン固定・キャッシュ・matrix実行のために `actions/setup-go` を使うのが標準。
- [x] `strategy.matrix` で複数バージョンを指定した場合、各バージョンは直列で動くか？並列で動くか？: 予想は並列（B）。実際も各matrixの組み合わせが個別のRunnerで並列実行される。
- [x] Matrixの1つが失敗した場合、実行中の他のMatrixジョブはどうなるか？（`fail-fast` の既定値）: 予想は「並列のためそのまま実行される（A）」。しかし実際はGitHub Actionsの既定値 `fail-fast: true` により、1つ失敗すると他ジョブが即座に自動キャンセルされる。完走させたい場合は `fail-fast: false` が必要。

## 壁打ちメモ

### 2026-09-24: Chapter 04開始
- `main` からブランチ `lesson/04-application-ci` を作成。

### 2026-09-24: コードのチェックアウトとGoランタイムの有無（予想と壁打ち）

1. **リポジトリのコード自動Clone**:
   - 予想: B（自動では配置されない）。
   - 実際: そのとおり。GitLab CI/CDではJob開始時にRunnerが自動でcloneするが、GitHub Actionsでは明示的に `actions/checkout@v4` を実行しないとコードが存在しない。
2. **Runnerのプリインストールと `actions/setup-go` の役割**:
   - ユーザーの鋭い指摘: 「今のubuntuってgoがデフォで入ってませんでしたっけ？」
   - 実際: **その通りで、GitHub-hosted Runner（ubuntu-latest等）にはGo、Node.js、Pythonなどが最初から入っている！**
   - なぜそれでも `actions/setup-go` を使うのか？
     1. **Runnerイメージ更新によるバージョン勝手上がり防止**: `ubuntu-latest` のデフォルトバージョンが変わるとCIが突然壊れる恐れがある。
     2. **プロジェクトのバージョン（`go.mod`等）とのピン留め**: 任意のバージョン（1.22.x等）を明示指定できる。
     3. **内蔵キャッシュ**: `cache: true`（既定値）によりモジュールキャッシュが自動化される。
     4. **Matrixテスト**: 後半で試す「複数Goバージョンでの並列テスト」には動的な切り替えが必要。

3. **`go-version-file: 'go.mod'` の実務的メリット**:
   - ユーザーの考察: 「`go-version-file` だけ（`go.mod` だけ）更新すればいいメリット」
   - 実際: まさにそのとおり。YAML側にバージョン番号（`1.22`）を直接ハードコードすると、Goのバージョンアップ時に `go.mod` と `.github/workflows/ci.yml` の2箇所を更新する必要があり、バージョン乖離の事故が起きやすい。`go-version-file: 'go.mod'` にしておけば Single Source of Truth（真実の単一情報源）を保つことができる。

4. **Job分割時の `checkout` と `setup-go` の重複記述**:
   - ユーザーの考察: 「B（全Jobで都度書く）が正解だと思うけど、理想はA（1回で引き継ぎ）がいい」
   - 実際:
     - 動作仕様としては **B**。Jobごとに完全に独立した新しい仮想マシンが起動するため、`lint`, `test`, `build` の全Jobで毎回 `actions/checkout` と `actions/setup-go` を書く必要がある。
     - **なぜこのトレードオフなのか？**:
       - メリット: Job間の依存がなく完全並列に起動できる（`lint` と `test` が互いを待たずに即時実行可能）。
       - デメリット: YAMLに同じセットアップステップが何度も登場し、Runner起動待ちやオーバーヘッドが生じる。
     - **実務での工夫**:
       - `setup-go` はRunner上のローカルキャッシュ（`/opt/hostedtoolcache/`）を使うため、2回目以降のダウンロード時間はほぼゼロ（数秒）。
       - 記述の重複を減らすには **Composite Action（複合Action）** や **Reusable Workflow** で一連の初期化を共通部品化する。
       - 規模が小さいCIなら、あえてJobを分割せず「1つのJob内にStepとして lint -> test -> build を並べる」方がトータル実行時間が短くシンプルになる場合もある。

5. **Matrix実行と `fail-fast` の設計思想**:
   - ユーザーの考察: 「Q1は並列（B）、Q2は並列で別マシンで動いているため最後まで実行される（A）」
   - 実際:
     - **Q1（並列実行）は大正解（B）**: `matrix` に指定したバージョンごとに別々のRunnerが割り当てられ、一斉に並列稼働する。
     - **Q2（失敗時の挙動）は実はB（即座に他ジョブもキャンセルされる）**:
       - GitLab CI/CD（最後まで走る）との最大のギャップ。
       - GitHub Actionsでは **`fail-fast: true` がデフォルト**。
       - 「マトリックスの1つが落ちた時点で全体のステータスは失敗になるため、Runner枠や課金時間を節約するために、他の実行中・待機中Matrixジョブも即座に強制終了する」というリソース節約思想によるもの。
       - 全バージョンの成否結果を一覧で確認したい場合は、明示的に **`strategy.fail-fast: false`** を書く必要がある。

## 試したことと結果

### 実験1: `actions/checkout` と `actions/setup-go` による最小限のテストJob

- PR: [#17](https://github.com/urchin-hat/study-github-action/pull/17)
- Workflow Run: [36015832080](https://github.com/urchin-hat/study-github-action/actions/runs/36015832080)
- 目的:
  - `actions/checkout@v4` でコードを取得する。
  - `actions/setup-go@v5` で `go-version-file: 'go.mod'` を読み込ませてGo環境を準備する。
  - `go test -v ./...` がGitHub Actions上で正常に実行されることを確認する。

#### 実行結果
- 実行時間: 27秒（すべて成功 ✅）
- `actions/setup-go` のログ:
  - `go.mod` に記述された `go 1.22` を検知し、自動的に `go1.22.12` をセットアップ（`GOROOT='/opt/hostedtoolcache/go/1.22.12/x64'`）。
- `go version` のログ:
  ```text
  === Go Version ===
  go version go1.22.12 linux/amd64
  ```
- `go test -v ./...` のログ:
  ```text
  === RUN   TestHealthHandler
  --- PASS: TestHealthHandler (0.00s)
  === RUN   TestHelloHandler
  === RUN   TestHelloHandler/default_world
  === RUN   TestHelloHandler/custom_name
  --- PASS: TestHelloHandler (0.00s)
      --- PASS: TestHelloHandler/default_world (0.00s)
      --- PASS: TestHelloHandler/custom_name (0.00s)
  PASS
  ok  	github.com/urchin-hat/study-github-action	0.003s
  ```

#### 分かったこと
- `actions/checkout@v4` により、リポジトリのコードがカレントディレクトリ（`/home/runner/work/study-github-action/study-github-action`）に正しく展開された。
- `actions/setup-go@v5` の `go-version-file: 'go.mod'` により、設定ファイルの二重管理をすることなく、コードと整合したGo環境が整った。
- 標準ライブラリのみの構成のため、外部依存関係のダウンロード（`go mod download`）がなく高速にテストが完了した。

### 実験2: `lint`, `test`, `build` のJob分離とDAG接続

- PR: [#17](https://github.com/urchin-hat/study-github-action/pull/17)
- Workflow Run: [36016518487](https://github.com/urchin-hat/study-github-action/actions/runs/36016518487)
- 目的:
  - CIの責務を `lint`（静的解析・フォーマット）、`test`（ユニットテスト）、`build`（バイナリ作成）の3つのJobに分離する。
  - `lint` と `test` が並列実行され、両方が成功した場合のみ `build` が実行されるDAG（`needs: [lint, test]`）を確認する。
  - 各Jobで独立して `checkout` と `setup-go` が実行される挙動を確認する。

#### 実行結果
- 各Jobのステータスと実行時間:
  - `Lint & Format`: ✅ **success** (23秒)
  - `Unit Test`: ✅ **success** (23秒)
  - `Build Binary`: ✅ **success** (20秒)
- 実行順序:
  - `Lint & Format` と `Unit Test` が同時に並列起動して並走した。
  - 両ジョブの完了後、`Build Binary` が起動してバイナリ `bin/study-server` が生成された。
- アノテーション警告（キャッシュの自動挙動）:
  - `setup-go` の内蔵キャッシュ機能により `go.sum` が探索されたが、今回は外部依存がないためスキップされた（`Restore cache failed: Dependencies file is not found in ... Supported file pattern: go.sum`）。

#### 分かったこと
- **DAGによる責務分離の実現**:
  - `needs: [lint, test]` により、GitLab CI/CDの `stages` を使わなくても「並列フェーズ → 後続ビルド」の綺麗なパイプラインが組めた。
- **ボイラープレートの必然性**:
  - 各Jobごとに `actions/checkout` と `actions/setup-go` を書く必要があったが、Runner上のツールキャッシュ（`/opt/hostedtoolcache`）のおかげで、セットアップ時間はわずか数秒でオーバーヘッドは小さかった。

### 実験3: `strategy.matrix` による複数Goバージョンの並列テスト

- PR: [#17](https://github.com/urchin-hat/study-github-action/pull/17)
- Workflow Run: [36016974349](https://github.com/urchin-hat/study-github-action/actions/runs/36016974349)
- 目的:
  - `strategy.matrix` で `go-version: ["1.21", "1.22", "1.23"]` を指定し、3バージョンのテストが別々のRunnerで並列実行されることを確認する。
  - 後続の `build` Job（`needs: [lint, test]`）が、3つのMatrixテストすべてが完了・成功するまで待機してから起動することを確認する。

#### 実行結果
- 各Jobのステータスと実行時間:
  - `Lint & Format`: ✅ **success** (22秒)
  - `Unit Test (Go 1.21)`: ✅ **success** (27秒)
  - `Unit Test (Go 1.22)`: ✅ **success** (27秒)
  - `Unit Test (Go 1.23)`: ✅ **success** (27秒)
  - `Build Binary`: ✅ **success** (24秒)
- 実行順序:
  - `Lint & Format` と、3つのMatrixテスト（Go 1.21 / 1.22 / 1.23）の計4ジョブが同時にRunnerへ割り当てられ、一斉に並列稼働した。
  - 4つの並列ジョブがすべて正常終了した後に、後続の `Build Binary` が起動してバイナリが生成された。

#### 分かったこと
- **Matrixによる並列展開**:
  - `strategy.matrix` に配列を渡すだけで、GitHub Actionsが自動的にRunnerを複数台プロビジョニングし、各バージョンを並列実行してくれる。
- **`needs` によるMatrix待機**:
  - 後続のJobで `needs: [test]` と指定した場合、Matrixの特定バージョンではなく「Matrixで展開されたすべてのジョブ」の完了・成功を待ってから実行されることが確認できた。

## つまずいた点

（実験を通して記録します）

## ブログへ残したい要点

（実験を通して記録します）

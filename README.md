# Learn by Transcript

Google Meetのtranscriptから英語表現を自動抽出し、学習用データベースに保存するCLIツール。

## 特徴

- Google Meetの文字起こしから単語・熟語・慣用表現を自動抽出
- Claude Sonnet 4.5による意味・優先度・カテゴリの自動判定
- 出現頻度による優先度の自動更新
- SQLiteによるローカルデータベース管理
- Anthropic Claude API / Vertex AI の切り替え可能

## 必要要件

- Go 1.21+
- Anthropic Claude APIキー または Vertex AIプロジェクト

## セットアップ

### 1. リポジトリをクローン

```bash
git clone https://github.com/mamyudapao/learn-by-transcript.git
cd learn-by-transcript
```

### 2. 依存関係をインストール

```bash
go mod download
```

### 3. 環境変数を設定

`.env.example`をコピーして`.env`を作成：

```bash
cp .env.example .env
```

#### オプションA: Anthropic Claude API（開発環境）

`.env`を編集：

```bash
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-your-api-key-here
MODEL_NAME=claude-sonnet-4-20250514
DB_PATH=./expressions.db
```

または、環境変数を直接export：

```bash
export ANTHROPIC_API_KEY="sk-ant-your-api-key-here"
export LLM_PROVIDER="anthropic"
```

#### オプションB: Vertex AI（本番環境）

**前提条件:**
- Google CloudプロジェクトでVertex AIが有効化されていること
- Claudeモデルへのアクセスが有効化されていること

**認証設定（ADC推奨）:**

```bash
# Application Default Credentials (ADC) でログイン
gcloud auth application-default login
```

または、サービスアカウントを使用：

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

`.env`を編集：

```bash
LLM_PROVIDER=vertexai
VERTEX_PROJECT_ID=your-gcp-project-id
VERTEX_LOCATION=us-central1
MODEL_NAME=claude-sonnet-4-20250514
DB_PATH=./expressions.db
```

### 4. ビルド

```bash
go build -o bin/extract ./cmd/extract
```

## 使い方

### 1. transcriptから表現を抽出

```bash
./bin/extract extract sample_transcript.txt
```

処理の流れ：
1. 単語抽出（プログラムで自動抽出）
2. 熟語・慣用表現抽出（LLMで抽出）
3. 優先度・意味・カテゴリ判定（LLMで判定）
4. SQLiteデータベースに保存
5. 出現頻度に応じて優先度を自動更新

### 2. CSV出力（Google Spreadsheets用）

```bash
./bin/extract export expressions.csv
```

出力されたCSVファイルをGoogle Spreadsheetsにインポート：
1. Google Sheetsを開く
2. File → Import → Upload
3. `expressions.csv`を選択

### 3. LLM接続テスト

```bash
./bin/extract test
```

### 4. データベース内の表現を一覧表示

```bash
./bin/extract list
```

## プロジェクト構成

```
learn-by-transcript/
├── cmd/
│   └── extract/          # CLIエントリーポイント
├── internal/
│   ├── config/           # 設定管理
│   ├── llm/              # LLMプロバイダー（Anthropic/Vertex AI）
│   ├── extractor/        # 表現抽出ロジック
│   ├── storage/          # SQLiteストレージ
│   ├── output/           # CSV出力
│   ├── service/          # メイン処理パイプライン
│   └── models/           # データモデル
├── pkg/
│   └── prompt/           # LLMプロンプトテンプレート
├── migrations/           # SQLiteスキーマ
└── README.md
```

## 開発ステータス

### ✅ 完了（フル機能実装済み）
- [x] プロジェクト構成
- [x] データモデル
- [x] LLMプロバイダーインターフェース（Anthropic Claude API / Vertex AI）
- [x] SQLiteストレージ（出現回数追跡、トリガー）
- [x] CLIコマンド（extract, export, test, list）
- [x] 単語抽出ロジック
- [x] 熟語・慣用表現抽出ロジック（LLM使用、バッチ処理対応）
- [x] 優先度判定ロジック（LLM使用、バッチ処理対応）
- [x] プロンプトテンプレート
- [x] transcriptファイル処理の統合
- [x] メインの処理フロー実装
- [x] 出現頻度による優先度自動更新
- [x] CSV出力機能（Google Spreadsheets対応）
- [x] Vertex AI実装（ADC/サービスアカウント対応）

### 🚧 今後の拡張案
- [ ] Notion API出力（直接登録）
- [ ] バッチ処理（複数ファイル一括処理）
- [ ] 表現の検索・フィルタリング機能
- [ ] 統計情報の表示
- [ ] Web UI

## ライセンス

MIT

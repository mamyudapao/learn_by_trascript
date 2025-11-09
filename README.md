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

`.env`を編集してAPIキーを設定：

```bash
# Anthropic Claude API（開発環境）
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

### 2. LLM接続テスト

```bash
./bin/extract test
```

### 3. データベース内の表現を一覧表示

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
│   ├── extractor/        # 表現抽出ロジック（未実装）
│   ├── storage/          # SQLiteストレージ
│   ├── output/           # 出力先（Notion/Sheets/CSV）（未実装）
│   └── models/           # データモデル
├── migrations/           # SQLiteスキーマ
└── README.md
```

## 開発ステータス

### ✅ 完了（MVPとして動作可能）
- [x] プロジェクト構成
- [x] データモデル
- [x] LLMプロバイダーインターフェース（Anthropic Claude API）
- [x] SQLiteストレージ（出現回数追跡、トリガー）
- [x] CLIコマンド（extract, test, list）
- [x] 単語抽出ロジック
- [x] 熟語・慣用表現抽出ロジック（LLM使用）
- [x] 優先度判定ロジック（LLM使用）
- [x] プロンプトテンプレート
- [x] transcriptファイル処理の統合
- [x] メインの処理フロー実装
- [x] 出現頻度による優先度自動更新

### 🚧 今後の拡張
- [ ] 出力先（Notion/Sheets/CSV）
- [ ] Vertex AI実装
- [ ] バッチ処理（複数ファイル一括処理）
- [ ] 表現の検索・フィルタリング機能
- [ ] 統計情報の表示

## ライセンス

MIT

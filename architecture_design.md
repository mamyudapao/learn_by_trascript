# アーキテクチャ設計 - 英語表現抽出CLI

## 概要

Google Meetのtranscriptから英語表現を抽出し、学習用データベースに保存するCLIツール。

## 技術スタック

- **言語**: Go
- **LLM**: Claude Sonnet 4.5
  - 開発環境: Anthropic Claude API（直接）
  - 本番環境: Vertex AI経由
- **データベース**: SQLite
- **出力先**: 切り替え可能（Notion / Google Sheets / CSV）

## プロジェクト構成

```
learn-by-transcript/
├── cmd/
│   └── extract/
│       └── main.go                 # CLIエントリーポイント
├── internal/
│   ├── config/
│   │   └── config.go              # 設定管理
│   ├── llm/
│   │   ├── provider.go            # LLMプロバイダーインターフェース
│   │   ├── anthropic.go           # Anthropic API実装
│   │   └── vertexai.go            # Vertex AI実装
│   ├── extractor/
│   │   ├── word_extractor.go      # 単語抽出（プログラム）
│   │   ├── phrase_extractor.go    # 熟語・慣用表現抽出（LLM）
│   │   └── prioritizer.go         # 優先度付け（LLM）
│   ├── storage/
│   │   ├── repository.go          # ストレージインターフェース
│   │   └── sqlite.go              # SQLite実装
│   ├── output/
│   │   ├── exporter.go            # 出力インターフェース
│   │   ├── csv.go                 # CSV出力
│   │   ├── notion.go              # Notion出力
│   │   └── sheets.go              # Google Sheets出力
│   └── models/
│       └── expression.go          # データモデル
├── pkg/
│   └── prompt/
│       └── templates.go           # プロンプトテンプレート
├── go.mod
├── go.sum
├── README.md
└── .env.example                   # 環境変数のサンプル
```

## データモデル

```go
package models

import "time"

// Expression は単語または熟語・慣用表現を表す
type Expression struct {
    ID              int       `db:"id"`
    Expression      string    `db:"expression"`        // 単語または熟語
    Type            string    `db:"type"`              // "word" または "phrase"
    Meaning         string    `db:"meaning"`           // 日本語の意味
    Context         string    `db:"context"`           // 使用された文脈
    Priority        int       `db:"priority"`          // 1(低) ~ 5(高)
    Category        string    `db:"category"`          // "engineering" / "business" / "casual"
    OccurrenceCount int       `db:"occurrence_count"`  // 出現回数
    FirstSeenAt     time.Time `db:"first_seen_at"`
    LastSeenAt      time.Time `db:"last_seen_at"`
    UpdatedAt       time.Time `db:"updated_at"`
}
```

## LLMプロバイダーインターフェース

```go
package llm

import "context"

// Provider はLLMプロバイダーの抽象インターフェース
type Provider interface {
    // Generate はプロンプトを送信して応答を取得
    Generate(ctx context.Context, prompt string) (string, error)

    // GetModelName は使用中のモデル名を取得
    GetModelName() string
}

// Config はLLMプロバイダーの設定
type Config struct {
    Type        string // "anthropic" or "vertexai"
    APIKey      string // Anthropic APIキー
    ProjectID   string // Vertex AI用のGCPプロジェクトID
    Location    string // Vertex AI用のロケーション
    Model       string // モデル名（例: "claude-sonnet-4-20250514"）
}

// NewProvider は設定に基づいて適切なプロバイダーを生成
func NewProvider(cfg Config) (Provider, error) {
    switch cfg.Type {
    case "anthropic":
        return NewAnthropicProvider(cfg.APIKey, cfg.Model)
    case "vertexai":
        return NewVertexAIProvider(cfg.ProjectID, cfg.Location, cfg.Model)
    default:
        return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
    }
}
```

## ストレージインターフェース

```go
package storage

import (
    "context"
    "github.com/yourusername/learn-by-transcript/internal/models"
)

// Repository はストレージの抽象インターフェース
type Repository interface {
    // SaveExpression は表現を保存（重複時はスキップ）
    SaveExpression(ctx context.Context, expr *models.Expression) error

    // GetAllExpressions はすべての表現を取得
    GetAllExpressions(ctx context.Context) ([]*models.Expression, error)

    // ExpressionExists は表現が既に存在するか確認
    ExpressionExists(ctx context.Context, expression string) (bool, error)

    // Close はリソースをクリーンアップ
    Close() error
}
```

## 出力インターフェース

```go
package output

import (
    "context"
    "github.com/yourusername/learn-by-transcript/internal/models"
)

// Exporter は出力先の抽象インターフェース
type Exporter interface {
    // Export は表現を出力先にエクスポート
    Export(ctx context.Context, expressions []*models.Expression) error
}
```

## 処理フロー

### 1. 全体フロー

```
transcript.txt
    ↓
[単語抽出（プログラム）]
    ↓
単語リスト
    ↓
[熟語・慣用表現抽出（LLM）]
    ↓
表現リスト（単語 + 熟語）
    ↓
[優先度付け（LLM）]
    ↓
優先度付き表現リスト
    ↓
[SQLite保存]
    ↓
[出力（CSV/Notion/Sheets）]
```

### 2. メイン処理（main.go）

```go
func main() {
    // 1. 設定読み込み
    cfg := config.Load()

    // 2. LLMプロバイダー初期化
    llmProvider, err := llm.NewProvider(cfg.LLM)

    // 3. ストレージ初期化
    repo, err := storage.NewSQLiteRepository(cfg.DBPath)
    defer repo.Close()

    // 4. transcriptファイル読み込み
    transcript, err := os.ReadFile(cfg.InputFile)

    // 5. 単語抽出（プログラム）
    wordExtractor := extractor.NewWordExtractor()
    words := wordExtractor.Extract(string(transcript))

    // 6. 熟語・慣用表現抽出（LLM）
    phraseExtractor := extractor.NewPhraseExtractor(llmProvider)
    phrases, err := phraseExtractor.Extract(ctx, string(transcript))

    // 7. 優先度付け（LLM）
    prioritizer := extractor.NewPrioritizer(llmProvider)
    allExpressions := append(words, phrases...)
    prioritizedExprs, err := prioritizer.Prioritize(ctx, allExpressions, string(transcript))

    // 8. SQLite保存（重複チェック含む）
    for _, expr := range prioritizedExprs {
        exists, _ := repo.ExpressionExists(ctx, expr.Expression)
        if !exists {
            repo.SaveExpression(ctx, expr)
        }
    }

    // 9. 出力（オプション）
    if cfg.OutputType != "" {
        exporter := output.NewExporter(cfg.OutputType, cfg.OutputConfig)
        saved, _ := repo.GetAllExpressions(ctx)
        exporter.Export(ctx, saved)
    }
}
```

## 設定ファイル（.env）

```bash
# LLMプロバイダー設定
LLM_PROVIDER=anthropic              # "anthropic" or "vertexai"
ANTHROPIC_API_KEY=sk-ant-xxx        # Anthropic APIキー
VERTEX_PROJECT_ID=my-gcp-project    # Vertex AI用
VERTEX_LOCATION=us-central1         # Vertex AI用
MODEL_NAME=claude-sonnet-4-20250514

# データベース
DB_PATH=./expressions.db

# 出力設定
OUTPUT_TYPE=csv                     # "csv", "notion", "sheets", ""（空=出力しない）
NOTION_API_KEY=xxx
NOTION_DATABASE_ID=xxx
SHEETS_CREDENTIALS_PATH=./credentials.json
SHEETS_SPREADSHEET_ID=xxx
```

## 次のステップ

- [ ] プロンプトテンプレートの設計
- [ ] SQLiteスキーマの詳細設計
- [ ] エラーハンドリング設計
- [ ] ロギング設計

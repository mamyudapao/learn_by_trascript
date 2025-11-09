package llm

import (
	"context"
	"fmt"
)

// Provider はLLMプロバイダーの抽象インターフェース
type Provider interface {
	// Generate はプロンプトを送信して応答を取得
	Generate(ctx context.Context, prompt string) (string, error)

	// GetModelName は使用中のモデル名を取得
	GetModelName() string
}

// Config はLLMプロバイダーの設定
type Config struct {
	Type      string // "anthropic" or "vertexai"
	APIKey    string // Anthropic APIキー
	ProjectID string // Vertex AI用のGCPプロジェクトID
	Location  string // Vertex AI用のロケーション
	Model     string // モデル名
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

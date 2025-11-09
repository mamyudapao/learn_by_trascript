package llm

import (
	"context"
	"fmt"
)

// VertexAIProvider はVertex AI経由でClaudeを利用するプロバイダー
type VertexAIProvider struct {
	projectID string
	location  string
	model     string
}

// NewVertexAIProvider は新しいVertexAIProviderを作成
func NewVertexAIProvider(projectID, location, model string) (*VertexAIProvider, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if location == "" {
		location = "us-central1" // デフォルト
	}
	if model == "" {
		model = "claude-sonnet-4-20250514" // デフォルト
	}

	return &VertexAIProvider{
		projectID: projectID,
		location:  location,
		model:     model,
	}, nil
}

// Generate はプロンプトを送信して応答を取得
func (p *VertexAIProvider) Generate(ctx context.Context, prompt string) (string, error) {
	// TODO: Vertex AI APIの実装
	// Google Cloud SDKを使用してVertex AI経由でClaudeを呼び出す
	return "", fmt.Errorf("Vertex AI provider not yet implemented")
}

// GetModelName は使用中のモデル名を取得
func (p *VertexAIProvider) GetModelName() string {
	return p.model
}

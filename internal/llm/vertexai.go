package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2/google"
)

// VertexAIProvider はVertex AI経由でClaudeを利用するプロバイダー
type VertexAIProvider struct {
	projectID  string
	location   string
	model      string
	httpClient *http.Client
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

	// Google Cloud認証（ADCまたはサービスアカウント）
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated client: %w (run 'gcloud auth application-default login' or set GOOGLE_APPLICATION_CREDENTIALS)", err)
	}

	return &VertexAIProvider{
		projectID:  projectID,
		location:   location,
		model:      model,
		httpClient: client,
	}, nil
}

// Generate はプロンプトを送信して応答を取得
func (p *VertexAIProvider) Generate(ctx context.Context, prompt string) (string, error) {
	// Vertex AI経由でClaudeを呼び出す
	// エンドポイント: https://{location}-aiplatform.googleapis.com/v1/projects/{project}/locations/{location}/publishers/anthropic/models/{model}:rawPredict
	url := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/anthropic/models/%s:rawPredict",
		p.location, p.projectID, p.location, p.model,
	)

	// Anthropic Messages APIと同じリクエスト形式
	requestBody := map[string]interface{}{
		"anthropic_version": "vertex-2023-10-16",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": 4096,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// リクエスト送信
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// レスポンスをパース
	var response struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return response.Content[0].Text, nil
}

// GetModelName は使用中のモデル名を取得
func (p *VertexAIProvider) GetModelName() string {
	return p.model
}

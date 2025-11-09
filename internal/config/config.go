package config

import (
	"fmt"
	"os"

	"github.com/mamyudapao/learn-by-transcript/internal/llm"
)

// Config はアプリケーション設定
type Config struct {
	LLM    llm.Config
	DBPath string
}

// Load は環境変数から設定を読み込む
func Load() (*Config, error) {
	llmType := getEnvOrDefault("LLM_PROVIDER", "anthropic")

	llmCfg := llm.Config{
		Type:      llmType,
		APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
		ProjectID: os.Getenv("VERTEX_PROJECT_ID"),
		Location:  getEnvOrDefault("VERTEX_LOCATION", "us-central1"),
		Model:     getEnvOrDefault("MODEL_NAME", "claude-sonnet-4-20250514"),
	}

	// 基本的なバリデーション
	if llmType == "anthropic" && llmCfg.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required when using anthropic provider")
	}
	if llmType == "vertexai" && llmCfg.ProjectID == "" {
		return nil, fmt.Errorf("VERTEX_PROJECT_ID is required when using vertexai provider")
	}

	cfg := &Config{
		LLM:    llmCfg,
		DBPath: getEnvOrDefault("DB_PATH", "./expressions.db"),
	}

	return cfg, nil
}

// LoadBasic はLLM設定なしで基本設定のみを読み込む（export, listコマンド用）
func LoadBasic() (*Config, error) {
	cfg := &Config{
		DBPath: getEnvOrDefault("DB_PATH", "./expressions.db"),
	}
	return cfg, nil
}

// getEnvOrDefault は環境変数を取得、なければデフォルト値を返す
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

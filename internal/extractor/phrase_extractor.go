package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mamyudapao/learn-by-transcript/internal/llm"
	"github.com/mamyudapao/learn-by-transcript/internal/models"
	"github.com/mamyudapao/learn-by-transcript/pkg/prompt"
)

// PhraseExtractor は熟語・慣用表現を抽出する
type PhraseExtractor struct {
	llmProvider llm.Provider
}

// NewPhraseExtractor は新しいPhraseExtractorを作成
func NewPhraseExtractor(provider llm.Provider) *PhraseExtractor {
	return &PhraseExtractor{
		llmProvider: provider,
	}
}

// Extract はテキストから熟語・慣用表現を抽出
func (e *PhraseExtractor) Extract(ctx context.Context, text string) ([]*models.Expression, error) {
	// プロンプト生成
	promptText := prompt.ExtractPhrasesPrompt(text)

	// LLM呼び出し
	response, err := e.llmProvider.Generate(ctx, promptText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// レスポンスをパース
	phrases, err := parsePhraseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return phrases, nil
}

// PhraseJSON はLLMからのレスポンスのJSON形式
type PhraseJSON struct {
	Phrase  string `json:"phrase"`
	Context string `json:"context"`
}

// parsePhraseResponse はLLMのレスポンスをパース
func parsePhraseResponse(response string) ([]*models.Expression, error) {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	expressions := make([]*models.Expression, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// JSONコードブロックのマーカーをスキップ
		if strings.HasPrefix(line, "```") {
			continue
		}

		var phraseData PhraseJSON
		if err := json.Unmarshal([]byte(line), &phraseData); err != nil {
			// パースエラーは警告して続行
			fmt.Printf("Warning: failed to parse line: %s (error: %v)\n", line, err)
			continue
		}

		if phraseData.Phrase == "" {
			continue
		}

		expr := &models.Expression{
			Expression: phraseData.Phrase,
			Type:       string(models.TypePhrase),
			Context:    phraseData.Context,
		}

		expressions = append(expressions, expr)
	}

	return expressions, nil
}

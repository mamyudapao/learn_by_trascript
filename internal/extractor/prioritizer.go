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

// Prioritizer は表現に優先度とカテゴリを付ける
type Prioritizer struct {
	llmProvider llm.Provider
}

// NewPrioritizer は新しいPrioritizerを作成
func NewPrioritizer(provider llm.Provider) *Prioritizer {
	return &Prioritizer{
		llmProvider: provider,
	}
}

// Prioritize は表現に優先度・カテゴリ・意味を付ける
func (p *Prioritizer) Prioritize(ctx context.Context, expressions []*models.Expression, transcript string) error {
	if len(expressions) == 0 {
		return nil
	}

	// バッチサイズ（一度に処理する表現数）
	const batchSize = 50

	// 全表現をマップ化（高速ルックアップ用）
	exprMap := make(map[string]*models.Expression)
	for _, expr := range expressions {
		exprMap[expr.Expression] = expr
	}

	// バッチごとに処理
	totalMatched := 0
	totalUnmatched := 0

	for i := 0; i < len(expressions); i += batchSize {
		end := i + batchSize
		if end > len(expressions) {
			end = len(expressions)
		}
		batch := expressions[i:end]

		fmt.Printf("  バッチ %d/%d を処理中 (%d-%d件目)...\n",
			i/batchSize+1, (len(expressions)+batchSize-1)/batchSize, i+1, end)

		// バッチの表現リストを作成
		exprList := make([]string, len(batch))
		for j, expr := range batch {
			exprList[j] = expr.Expression
		}

		// プロンプト生成
		promptText := prompt.PrioritizeExpressionsPrompt(exprList, transcript)

		// LLM呼び出し
		response, err := p.llmProvider.Generate(ctx, promptText)
		if err != nil {
			return fmt.Errorf("failed to generate response for batch %d: %w", i/batchSize+1, err)
		}

		// レスポンスをパース
		priorityMap, err := parsePriorityResponse(response)
		if err != nil {
			return fmt.Errorf("failed to parse response for batch %d: %w", i/batchSize+1, err)
		}

		// バッチ内の各表現に優先度・カテゴリ・意味を設定
		batchMatched := 0
		batchUnmatched := 0
		for _, expr := range batch {
			if data, ok := priorityMap[expr.Expression]; ok {
				expr.Priority = data.Priority
				expr.Category = data.Category
				expr.Meaning = data.Meaning
				batchMatched++
			} else {
				// デフォルト値
				expr.Priority = 3
				expr.Category = string(models.CategoryBusiness)
				expr.Meaning = ""
				batchUnmatched++
			}
		}

		totalMatched += batchMatched
		totalUnmatched += batchUnmatched
		fmt.Printf("    マッチ: %d/%d\n", batchMatched, len(batch))
	}

	fmt.Printf("  全体マッチ: %d/%d expressions\n", totalMatched, len(expressions))
	if totalUnmatched > 0 {
		fmt.Printf("  警告: %d個の表現にデフォルト値を設定しました\n", totalUnmatched)
	}

	return nil
}

// PriorityJSON はLLMからのレスポンスのJSON形式
type PriorityJSON struct {
	Expression string `json:"expression"`
	Meaning    string `json:"meaning"`
	Priority   int    `json:"priority"`
	Category   string `json:"category"`
}

// parsePriorityResponse はLLMのレスポンスをパース
func parsePriorityResponse(response string) (map[string]PriorityJSON, error) {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	result := make(map[string]PriorityJSON)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// JSONコードブロックのマーカーをスキップ
		if strings.HasPrefix(line, "```") {
			continue
		}

		var data PriorityJSON
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			// パースエラーは警告して続行
			fmt.Printf("Warning: failed to parse line: %s (error: %v)\n", line, err)
			continue
		}

		if data.Expression == "" {
			continue
		}

		result[data.Expression] = data
	}

	return result, nil
}

// UpdatePriorityBasedOnOccurrence は出現回数に基づいて優先度を更新
func UpdatePriorityBasedOnOccurrence(basePriority int, occurrenceCount int) int {
	boost := 0
	if occurrenceCount >= 5 {
		boost = 2 // 5回以上出現 → +2
	} else if occurrenceCount >= 3 {
		boost = 1 // 3回以上出現 → +1
	}

	newPriority := basePriority + boost
	if newPriority > 5 {
		newPriority = 5 // 最大5
	}

	return newPriority
}

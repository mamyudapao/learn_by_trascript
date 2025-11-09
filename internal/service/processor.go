package service

import (
	"context"
	"fmt"

	"github.com/mamyudapao/learn-by-transcript/internal/extractor"
	"github.com/mamyudapao/learn-by-transcript/internal/llm"
	"github.com/mamyudapao/learn-by-transcript/internal/storage"
)

// TranscriptProcessor はtranscript処理のメインロジック
type TranscriptProcessor struct {
	wordExtractor   *extractor.WordExtractor
	phraseExtractor *extractor.PhraseExtractor
	prioritizer     *extractor.Prioritizer
	repository      storage.Repository
}

// NewTranscriptProcessor は新しいTranscriptProcessorを作成
func NewTranscriptProcessor(provider llm.Provider, repo storage.Repository) *TranscriptProcessor {
	return &TranscriptProcessor{
		wordExtractor:   extractor.NewWordExtractor(),
		phraseExtractor: extractor.NewPhraseExtractor(provider),
		prioritizer:     extractor.NewPrioritizer(provider),
		repository:      repo,
	}
}

// ProcessResult は処理結果
type ProcessResult struct {
	TotalExpressions int
	NewExpressions   int
	UpdatedPriority  int
}

// Process はtranscriptを処理して表現を抽出・保存
func (p *TranscriptProcessor) Process(ctx context.Context, transcript string) (*ProcessResult, error) {
	result := &ProcessResult{}

	fmt.Println("Step 1: 単語抽出中...")
	// 1. 単語抽出
	words := p.wordExtractor.ExtractWithContext(transcript)
	fmt.Printf("  抽出された単語: %d個\n", len(words))

	fmt.Println("\nStep 2: 熟語・慣用表現抽出中（LLM使用）...")
	// 2. 熟語・慣用表現抽出
	phrases, err := p.phraseExtractor.Extract(ctx, transcript)
	if err != nil {
		return nil, fmt.Errorf("failed to extract phrases: %w", err)
	}
	fmt.Printf("  抽出された熟語: %d個\n", len(phrases))

	// 3. 全表現をマージ
	allExpressions := append(words, phrases...)
	result.TotalExpressions = len(allExpressions)

	fmt.Println("\nStep 3: 優先度・意味・カテゴリ判定中（LLM使用）...")
	// 4. 優先度・意味・カテゴリ判定
	if err := p.prioritizer.Prioritize(ctx, allExpressions, transcript); err != nil {
		return nil, fmt.Errorf("failed to prioritize expressions: %w", err)
	}
	fmt.Println("  判定完了")

	fmt.Println("\nStep 4: データベースに保存中...")
	// 5. データベースに保存（重複チェック含む）
	for _, expr := range allExpressions {
		// 既存チェック
		exists, err := p.repository.ExpressionExists(ctx, expr.Expression)
		if err != nil {
			return nil, fmt.Errorf("failed to check existence: %w", err)
		}

		if exists {
			// 既存の場合: 出現履歴を追加
			existing, err := p.repository.GetExpression(ctx, expr.Expression)
			if err != nil {
				return nil, fmt.Errorf("failed to get existing expression: %w", err)
			}

			// 出現履歴追加
			if err := p.repository.AddOccurrence(ctx, existing.ID, expr.Context); err != nil {
				return nil, fmt.Errorf("failed to add occurrence: %w", err)
			}

			// 出現回数を取得（トリガーで更新されている）
			updated, err := p.repository.GetExpression(ctx, expr.Expression)
			if err != nil {
				return nil, fmt.Errorf("failed to get updated expression: %w", err)
			}

			// 出現頻度に基づいて優先度を更新
			newPriority := extractor.UpdatePriorityBasedOnOccurrence(updated.Priority, updated.OccurrenceCount)
			if newPriority != updated.Priority {
				if err := p.repository.UpdatePriority(ctx, updated.ID, newPriority); err != nil {
					return nil, fmt.Errorf("failed to update priority: %w", err)
				}
				result.UpdatedPriority++
				fmt.Printf("  '%s' の優先度を更新: %d → %d (出現回数: %d)\n",
					expr.Expression, updated.Priority, newPriority, updated.OccurrenceCount)
			}
		} else {
			// 新規の場合: 保存
			if err := p.repository.SaveExpression(ctx, expr); err != nil {
				return nil, fmt.Errorf("failed to save expression: %w", err)
			}

			// 最初の出現履歴を追加
			if err := p.repository.AddOccurrence(ctx, expr.ID, expr.Context); err != nil {
				return nil, fmt.Errorf("failed to add first occurrence: %w", err)
			}

			result.NewExpressions++
			fmt.Printf("  新規: '%s' (優先度: %d, カテゴリ: %s)\n",
				expr.Expression, expr.Priority, expr.Category)
		}
	}

	return result, nil
}

package output

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/mamyudapao/learn-by-transcript/internal/models"
	"github.com/mamyudapao/learn-by-transcript/internal/storage"
)

// CSVExporter はCSVファイルに出力する
type CSVExporter struct {
	repository storage.Repository
}

// NewCSVExporter は新しいCSVExporterを作成
func NewCSVExporter(repo storage.Repository) *CSVExporter {
	return &CSVExporter{
		repository: repo,
	}
}

// Export はデータベースの表現をCSVファイルに出力
func (e *CSVExporter) Export(ctx context.Context, outputPath string) error {
	// データベースからすべての表現を取得
	expressions, err := e.repository.ListExpressions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list expressions: %w", err)
	}

	// CSVファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// ヘッダー行
	header := []string{
		"Expression",
		"Type",
		"Meaning",
		"Priority",
		"Category",
		"OccurrenceCount",
		"Context",
		"FirstSeenAt",
		"LastSeenAt",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// データ行
	for _, expr := range expressions {
		// 最新のcontextを取得
		context := ""
		occurrences, err := e.repository.GetOccurrences(ctx, expr.ID)
		if err == nil && len(occurrences) > 0 {
			// 最新のoccurrenceのcontextを使用
			context = occurrences[len(occurrences)-1].Context
		}

		row := []string{
			expr.Expression,
			expr.Type,
			expr.Meaning,
			fmt.Sprintf("%d", expr.Priority),
			expr.Category,
			fmt.Sprintf("%d", expr.OccurrenceCount),
			context,
			expr.FirstSeenAt.Format("2006-01-02 15:04:05"),
			expr.LastSeenAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// ExportWithOptions はオプション付きでCSVに出力
type ExportOptions struct {
	MinPriority    int    // 最小優先度（フィルタリング）
	Category       string // カテゴリでフィルタ（空文字列ならすべて）
	SortBy         string // "priority", "occurrence", "expression"
	IncludeContext bool   // contextを含めるか
}

// ExportWithFilter はフィルタリング・並び替えを適用してCSV出力
func (e *CSVExporter) ExportWithFilter(ctx context.Context, outputPath string, opts ExportOptions) error {
	// データベースからすべての表現を取得
	allExpressions, err := e.repository.ListExpressions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list expressions: %w", err)
	}

	// フィルタリング
	var expressions []*models.Expression
	for _, expr := range allExpressions {
		// 優先度フィルタ
		if opts.MinPriority > 0 && expr.Priority < opts.MinPriority {
			continue
		}
		// カテゴリフィルタ
		if opts.Category != "" && expr.Category != opts.Category {
			continue
		}
		expressions = append(expressions, expr)
	}

	// CSVファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// ヘッダー行
	header := []string{
		"Expression",
		"Type",
		"Meaning",
		"Priority",
		"Category",
		"OccurrenceCount",
	}
	if opts.IncludeContext {
		header = append(header, "Context")
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// データ行
	for _, expr := range expressions {
		row := []string{
			expr.Expression,
			expr.Type,
			expr.Meaning,
			fmt.Sprintf("%d", expr.Priority),
			expr.Category,
			fmt.Sprintf("%d", expr.OccurrenceCount),
		}

		if opts.IncludeContext {
			context := ""
			occurrences, err := e.repository.GetOccurrences(ctx, expr.ID)
			if err == nil && len(occurrences) > 0 {
				context = occurrences[len(occurrences)-1].Context
			}
			row = append(row, context)
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

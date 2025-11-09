package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mamyudapao/learn-by-transcript/internal/config"
	"github.com/mamyudapao/learn-by-transcript/internal/llm"
	"github.com/mamyudapao/learn-by-transcript/internal/output"
	"github.com/mamyudapao/learn-by-transcript/internal/service"
	"github.com/mamyudapao/learn-by-transcript/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// コマンドライン引数チェック
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <command> [args]\n\nCommands:\n  extract <file> - Extract expressions from transcript file\n  export <output-file> - Export expressions to CSV file\n  test - Test LLM connection\n  list - List all expressions", os.Args[0])
	}

	command := os.Args[1]

	// 設定読み込み（コマンドに応じて使い分け）
	var cfg *config.Config
	var err error
	if command == "export" || command == "list" {
		// LLM不要なコマンドは基本設定のみ
		cfg, err = config.LoadBasic()
	} else {
		// LLM必要なコマンドは完全な設定
		cfg, err = config.Load()
	}
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// ストレージ初期化（export, listで必要）
	var repo storage.Repository
	if command == "export" || command == "list" || command == "extract" {
		repo, err = storage.NewSQLiteRepository(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}
		defer repo.Close()
		fmt.Printf("Database: %s\n", cfg.DBPath)
	}

	// LLMプロバイダー初期化（extract, testで必要）
	var provider llm.Provider
	if command == "extract" || command == "test" {
		provider, err = llm.NewProvider(cfg.LLM)
		if err != nil {
			return fmt.Errorf("failed to create LLM provider: %w", err)
		}
		fmt.Printf("Using LLM: %s (%s)\n", cfg.LLM.Type, provider.GetModelName())
	}

	switch command {
	case "extract":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: %s extract <transcript-file>", os.Args[0])
		}
		return extractFromFile(ctx, provider, repo, os.Args[2])
	case "export":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: %s export <output-file>", os.Args[0])
		}
		return exportToCSV(ctx, repo, os.Args[2])
	case "test":
		return testLLM(ctx, provider)
	case "list":
		return listExpressions(ctx, repo)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func extractFromFile(ctx context.Context, provider llm.Provider, repo storage.Repository, filePath string) error {
	fmt.Printf("\nProcessing transcript file: %s\n\n", filePath)

	// ファイル読み込み
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	transcript := string(content)
	fmt.Printf("Transcript length: %d characters\n\n", len(transcript))

	// プロセッサ作成
	processor := service.NewTranscriptProcessor(provider, repo)

	// 処理実行
	result, err := processor.Process(ctx, transcript)
	if err != nil {
		return fmt.Errorf("failed to process transcript: %w", err)
	}

	// 結果表示
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("処理完了")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("抽出した表現: %d個\n", result.TotalExpressions)
	fmt.Printf("新規登録: %d個\n", result.NewExpressions)
	fmt.Printf("優先度更新: %d個\n", result.UpdatedPriority)
	fmt.Println(strings.Repeat("=", 50))

	return nil
}

func testLLM(ctx context.Context, provider llm.Provider) error {
	fmt.Println("\nTesting LLM connection...")

	response, err := provider.Generate(ctx, "Say 'Hello, World!' in Japanese.")
	if err != nil {
		return fmt.Errorf("LLM test failed: %w", err)
	}

	fmt.Printf("Response: %s\n", response)
	return nil
}

func listExpressions(ctx context.Context, repo storage.Repository) error {
	fmt.Println("\nListing expressions...")

	expressions, err := repo.GetAllExpressions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expressions: %w", err)
	}

	if len(expressions) == 0 {
		fmt.Println("No expressions found.")
		return nil
	}

	fmt.Printf("Found %d expression(s):\n\n", len(expressions))
	for _, expr := range expressions {
		fmt.Printf("- %s (%s)\n", expr.Expression, expr.Type)
		fmt.Printf("  Meaning: %s\n", expr.Meaning)
		fmt.Printf("  Priority: %d, Occurrences: %d\n", expr.Priority, expr.OccurrenceCount)
		fmt.Printf("  Category: %s\n\n", expr.Category)
	}

	return nil
}

func exportToCSV(ctx context.Context, repo storage.Repository, outputPath string) error {
	fmt.Printf("\nExporting expressions to CSV: %s\n\n", outputPath)

	// CSV Exporter作成
	exporter := output.NewCSVExporter(repo)

	// エクスポート実行
	if err := exporter.Export(ctx, outputPath); err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	// 件数確認
	expressions, err := repo.ListExpressions(ctx)
	if err != nil {
		return fmt.Errorf("failed to count expressions: %w", err)
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("エクスポート完了")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("出力ファイル: %s\n", outputPath)
	fmt.Printf("エクスポート件数: %d個\n", len(expressions))
	fmt.Println(strings.Repeat("=", 50))

	return nil
}

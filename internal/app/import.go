package app

import (
	"context"
	"fmt"
	"sample/config"
	"sample/internal/infrastructure/analyzer"
	"sample/internal/infrastructure/file"
	"sample/internal/infrastructure/ollama"
	"sample/internal/infrastructure/postgres"
	usecase "sample/internal/usecase/import"
	"sample/setup"
	// Kagomeのインポート
)

func RunImport(ctx context.Context, cfg *config.Config) {
	// DBセットアップ
	db, err := setup.SetupApp(cfg)
	if err != nil {
		panic(fmt.Errorf("DB接続エラー: %v", err))
	}
	defer db.Close()

	// =================================================
	// 依存性注入
	// =================================================
	// infrastructure層を使用するために必要な設定を注入
	docRepo := postgres.NewImportRepository(db)
	llmService := ollama.NewLLMService(cfg.OllamaURL)
	fileService := file.NewFileService()
	analyzeService := analyzer.NewAnalyzerService()
	// Usecase層に具体的な実装を注入
	importUsecase := usecase.NewImportUsecase(docRepo, fileService, llmService, analyzeService)
	// usecase実行
	err = importUsecase.ImportFromDirectory(ctx, cfg.SourceDir, cfg.ProcessedDir)
	if err != nil {
		panic(fmt.Errorf("取り込みエラー: %v", err))
	}
}

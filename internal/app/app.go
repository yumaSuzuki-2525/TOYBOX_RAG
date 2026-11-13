package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sample/config"
	"sample/internal/infrastructure/ollama"
	"sample/internal/infrastructure/postgres"
	usecase "sample/internal/usecase/cli"
	"sample/setup"
	"strings"
)

func RunApp(ctx context.Context, cfg *config.Config) {
	setup.SetupApp(cfg)
	db, err := setup.SetupApp(cfg)
	if err != nil {
		panic(fmt.Errorf("DB接続エラー: %v", err))
	}
	defer db.Close()

	// =================================================
	// 依存性注入
	// =================================================
	// infrastructure層を使用するために必要な設定を注入
	docRepo := postgres.NewDocumentRepository(db)
	llmService := ollama.NewLLMService(cfg.OllamaURL)
	// Usecase層に具体的な実装を注入
	ragUsecase := usecase.NewRagUsecase(docRepo, llmService)

	// ====================================================
	// CLIで会話開始
	// ====================================================
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("=== チャット開始 ===")
	for {
		fmt.Print("\n質問 > ")
		if !scanner.Scan() {
			break // Ctrl+C や EOF で終了
		}

		userQuestion := strings.TrimSpace(scanner.Text())
		if userQuestion == "" {
			continue // 空エンターは無視
		}
		if userQuestion == "exit" || userQuestion == "quit" {
			break
		}

		// RAGの検索＆生成関数を呼び出し
		fmt.Printf("「%s」について検索中...\n", userQuestion)
		// 質問実行
		err = ragUsecase.RagAnswer(ctx, userQuestion)
		if err != nil {
			fmt.Printf("処理中にエラーが発生しました: %v\n", err)
		}
	}

}

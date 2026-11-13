package main

import (
	"context"
	"log"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"sample/config"
	"sample/internal/app"

	_ "github.com/lib/pq"
)

// 〼 Embedding（エンベディング：埋め込み）
//
//	意味： 文章をAIが理解できる「数値の配列（ベクトル）」に変換して空間に配置（埋め込み）すること。
//	役割： 「箱の回し方」という曖昧な言葉を計算可能なデータに変更。
//
// 〼 Retrieval（リトリーバル：検索・抽出）
//
//	意味： 蓄積されたデータの中から、必要な情報を引っぱり出してくること。RAGの最初の文字「R」です。
//	役割： ベクトル化された質問をDB（Postgres）で検索し、意味が最も近い過去のドキュメントを取得。
//
// 〼 Generation（ジェネレーション：生成）
//
//	意味： 文章を作り出すこと。RAGの最後の文字「G」です。
//	役割： LLM（Llama 3）が、検索してきた資料を読み込みながら、人間にわかりやすい回答を作成します。
//	OllamaのAPIエンドポイント

// 〼 OllamaとLlama 3の関係性
// Ollama：本体
//
//	AIを動かすための「実行環境（エンジン）」です。これ自体には知識はない。
//
// Llama 3：ソフト
//
//	世界中のデータを取り込んで作られた「AIのモデル」。

func main() {
	ctx := context.Background()
	// ========================================
	// CPUプロファイル収集。make pprof_cpu
	// ========================================
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("CPUプロファイル作成エラー:", err)
	}
	defer cpuFile.Close()

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatal("CPUプロファイル開始エラー:", err)
	}
	defer pprof.StopCPUProfile()

	// ========================================
	// Trace収集。 make pprof_trace
	// ========================================
	traceFile, err := os.Create("trace.out")
	if err != nil {
		log.Fatal("Traceファイル作成エラー:", err)
	}
	defer traceFile.Close()

	if err := trace.Start(traceFile); err != nil {
		log.Fatal("Trace開始エラー:", err)
	}
	defer trace.Stop()
	// ========================================
	// セットアップ
	// ========================================
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	// defer cfg.DB.Close()
	app.RunApp(ctx, cfg)
}

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
	app.RunImport(ctx, cfg)
}

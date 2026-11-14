// Package usecase implements ...
package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"sample/internal/domain"
	"strings"
)

//go:generate moq -out moq_test.go -pkg usecase ../../domain ImportRepository FileService ImportService AnalyzerService

type ImportUsecase struct {
	repo domain.ImportRepository
	file domain.FileService
	imp  domain.ImportService
	alz  domain.AnalyzerService
}

func NewImportUsecase(repo domain.ImportRepository, file domain.FileService, imp domain.ImportService,
	alz domain.AnalyzerService) *ImportUsecase {
	return &ImportUsecase{
		repo: repo,
		file: file,
		imp:  imp,
		alz:  alz,
	}
}

// ImportFromDirectory は指定フォルダ配下を探索し、処理が成功したものを別フォルダに移動します
// sourceDir: 読み込み元の親フォルダ (例: "./public/docs")
// processedDir: 処理済みの退避先フォルダ (例: "./public/processed")
func (u *ImportUsecase) ImportFromDirectory(ctx context.Context, sourceDir, processedDir string) error {
	err := filepath.WalkDir(sourceDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("パスのアクセスエラー %s: %w", currentPath, err)
		}
		// ディレクトリ自身はスキップ
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(currentPath))
		var content string

		// ファイルの読み込み・テキスト抽出
		switch ext {
		case ".txt", ".md", ".csv", ".json":
			content, err = u.file.ReadPlainFile(currentPath)
			if err != nil {
				fmt.Printf("読み込み失敗 (%s): %v\n", currentPath, err)
				return nil // 失敗した場合は移動せず、次のファイルへ
			}
		case ".pdf":
			// PDF抽出ロジック（エラー時は return nil して次へ）
			content, err = u.file.ReadPdfFile(currentPath)
			if err != nil {
				fmt.Printf("読み込み失敗 (%s): %v\n", currentPath, err)
				return nil // 失敗した場合は移動せず、次のファイルへ
			}

		default:
			fmt.Printf("スキップ(非対応): %s\n", currentPath)
			return nil
		}

		// 2. テキストが抽出できたらRAGへの取り込み（ベクトル化と保存）を実行
		if strings.TrimSpace(content) != "" {
			chunks, err := u.alz.ChunkTextBySemantics(content, 300)
			// chunks := chunkText(content, 400, 50)
			if err != nil {
				fmt.Printf("チャンク分割エラー: %v\n", err)
				return nil
			}
			// 各チャンクをベクトル化してDBに保存
			for i, chunk := range chunks {
				// // 冒頭40文字をプレビュー
				// preview := chunk
				// if len([]rune(chunk)) > 40 {
				// 	preview = string([]rune(chunk)[:40]) + "..."
				// }
				// // 改行が混ざると見にくい為取り除く
				// fmt.Printf("%d%dを処理(内容%s)\n", i+1, len(chunks), preview)

				// ベクトル化(Embbedding)
				vector, err := u.imp.GetEmbedding(chunk)
				if err != nil {
					fmt.Printf("チャンク%dのベクトル化に失敗:%v\n", i+1, err)
					continue
				}

				// JSONエンコードを使って確実にカンマ区切りの配列文字列にする
				vectorBytes, _ := json.Marshal(vector)
				vectorStr := string(vectorBytes)

				// DBへ保存
				err = u.repo.InsertData(ctx, chunk, vectorStr)
				if err != nil {
					fmt.Printf("チャンク %d のDB保存に失敗: %v\n", i+1, err)
					continue
				}
			}
			fmt.Printf("[取り込み完了] %s\n", currentPath)

			// 取り込みに成功した場合のみ、別フォルダへ階層を維持して移動
			if moveErr := u.file.MoveToProcessed(sourceDir, processedDir, currentPath); moveErr != nil {
				// DB保存は成功したがファイルの移動に失敗した場合のログ
				fmt.Printf("ファイル移動エラー (%s): %v\n", currentPath, moveErr)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("ディレクトリ探索エラー: %w", err)
	}

	return nil
}

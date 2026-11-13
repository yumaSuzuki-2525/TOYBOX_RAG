package file

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// LLMService は domain.LLMService を満たすOllama用の実装
type FileService struct{}

// NewFileService は OllamaのURLを受け取ってサービスを作成
func NewFileService() *FileService {
	return &FileService{}
}

// Plainファイル読み込み
func (s *FileService) ReadPlainFile(path string) (string, error) {
	bytes, readErr := os.ReadFile(path)
	if readErr != nil {
		return "", fmt.Errorf("ファイルの読み込みに失敗: %v", readErr)
	}
	return string(bytes), readErr
}

// Pdfファイル読み込み
func (s *FileService) ReadPdfFile(path string) (string, error) {
	// 第1引数: コマンド名
	// 第2〜3引数: 日本語の文字化けを防ぐUTF-8指定
	// 第4引数: 対象のPDFファイルパス
	// 第5引数: "-" は「テキストファイルに保存せず、直接結果を返せ」という指定
	cmd := exec.Command("pdftotext", "-enc", "UTF-8", path, "-")
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// stderr.String() をエラーメッセージに含める
		return "", fmt.Errorf("pdftotextの実行に失敗: %w, 詳細: %s", err, stderr.String())
	}

	return out.String(), nil
}

// MoveToProcessedはファイル移動します。
func (s *FileService) MoveToProcessed(sourceDir, processedDir, currentPath string) error {
	// 読み込み元フォルダを基準とした「相対パス」を計算
	// 例: sourceDir="./docs", currentPath="./docs/hr/rules.md" -> relPath="hr/rules.md"
	relPath, err := filepath.Rel(sourceDir, currentPath)
	if err != nil {
		return fmt.Errorf("相対パスの計算に失敗: %w", err)
	}

	// 移動先のフルパスを組み立て
	// 例: processedDir="./processed", relPath="hr/rules.md" -> destPath="./processed/hr/rules.md"
	destPath := filepath.Join(processedDir, relPath)

	// 移動先の親フォルダ（ディレクトリ）が存在しない場合は作成（階層ごと一気に作成）
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("退避先ディレクトリの作成に失敗: %w", err)
	}

	// ファイルを移動（名前変更）
	if err := os.Rename(currentPath, destPath); err != nil {
		return fmt.Errorf("ファイルの移動に失敗: %w", err)
	}

	fmt.Printf("退避完了 -> %s\n", destPath)
	return nil
}

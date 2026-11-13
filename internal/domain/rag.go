package domain

import "context"

// EmbeddingServiceは、ベクトル化だけを行うインターフェース
type EmbeddingService interface {
	GetEmbedding(text string) ([]float64, error)
}

// TextGeneratorは、テキスト生成だけを行うインターフェース
type TextGenerator interface {
	GenerateAnswer(prompt string) (string, error)
}
// ChunkTextBySemanticsは、形態素解析でチャンクを取得
type AnalyzerService interface {
	ChunkTextBySemantics(text string, maxRunes int) ([]string, error)
}

// LLMServiceはAIに対する操作
type LLMService interface {
	EmbeddingService
	TextGenerator
}

// ImportServiceはファイル取り込み時AIに対する操作
type ImportService interface {
	EmbeddingService
}

// FileServiceは、ファイルの読み、移動の操作をします。
type FileService interface {
	// Plainファイル読み込み
	ReadPlainFile(path string) (string, error)
	// Pdfファイル読み込み
	ReadPdfFile(path string) (string, error)
	// MoveToProcessedはファイル移動します。
	MoveToProcessed(sourceDir, processedDir, currentPath string) error
}

// DocumentRepositoryはDBに対する操作
type DocumentRepository interface {
	SearchSimilar(ctx context.Context, vector []float64) (string, error)
}

// ImpoertRepositoryはDBに対する操作
type ImpoertRepository interface {
	InsertData(ctx context.Context, chunk string, vectorStr string) error
}

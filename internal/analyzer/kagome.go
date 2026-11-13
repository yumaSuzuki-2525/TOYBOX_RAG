package analyzer

import (
	"fmt"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// LLMService は domain.LLMService を満たすOllama用の実装
type AnalyzerService struct{}

// NewAnalyzerService は OllamaのURLを受け取ってサービスを作ります
func NewAnalyzerService() *AnalyzerService {
	return &AnalyzerService{}
}

func (s *AnalyzerService) ChunkTextBySemantics(text string, maxRunes int) ([]string, error) {
	// IPA辞書を指定してトークナイザ（解析器）を生成
	// OmitBosEos は文の開始/終了を示す特殊なダミートークンを省略するオプションです
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("形態素解析器の初期化エラー: %w", err)
	}

	// テキストを形態素（単語）に分割
	tokens := t.Tokenize(text)

	var chunks []string
	var currentChunk strings.Builder
	var currentLength int

	for _, token := range tokens {
		word := token.Surface
		wordLen := len([]rune(word)) // バイト数ではなく実際の文字数を取得

		// 次の単語を足すと最大文字数を超える場合、現在の塊を保存してリセット
		// 単語の途中でスライスされないため、文字化けや意味の破損が起きない
		if currentLength+wordLen > maxRunes && currentLength > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
			currentLength = 0
		}

		currentChunk.WriteString(word)
		currentLength += wordLen

		// 2. 句点「。」が来たら、文字数に余裕があっても文脈の区切りとして強制的にチャンクを分割する
		// これにより、RAGのAIが「1つの完全な文章」として理解しやすくなる
		if word == "。" {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
			currentLength = 0
		}
	}

	// ループ終了後に残っているテキストがあれば最後のチャンクとして追加
	if currentLength > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	// 空のチャンクを除外するフィルタリング
	var validChunks []string
	for _, c := range chunks {
		if c != "" {
			validChunks = append(validChunks, c)
		}
	}

	return validChunks, nil
}

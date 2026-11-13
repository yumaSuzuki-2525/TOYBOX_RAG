package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMService は domain.LLMService を満たすOllama用の実装
type LLMService struct {
	baseURL string
}

// NewLLMService は OllamaのURLを受け取ってサービスを作成
func NewLLMService(baseURL string) *LLMService {
	return &LLMService{baseURL: baseURL}
}

// GetEmbedding は質問文をベクトルに変換します。
func (s *LLMService) GetEmbedding(text string) ([]float64, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"model":  "nomic-embed-text",
		"prompt": text,
	})

	resp, err := http.Post(
		s.baseURL+"/embeddings",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var res struct {
		Embedding []float64 `json:"embedding"`
	}
	json.Unmarshal(body, &res)

	return res.Embedding, nil
}

// GenerateAnswer は LLMが与えられたプロンプトから回答します。
func (s *LLMService) GenerateAnswer(prompt string) (string, error) {
	start := time.Now()
	defer func() {
		fmt.Printf("generateAnswer: %s\n", time.Since(start))
	}()

	reqBody, _ := json.Marshal(map[string]any{
		"model":  "llama3",
		"prompt": prompt,
		"stream": false,
	})

	resp, err := http.Post(
		s.baseURL+"/generate",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var res struct {
		Response string `json:"response"`
	}
	json.Unmarshal(body, &res)

	return res.Response, nil
}

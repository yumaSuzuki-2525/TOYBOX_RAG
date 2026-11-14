// Package usecase implements ...
package usecase

import (
	"context"
	"fmt"
	"sample/internal/domain"
)

type RagUsecase struct {
	repo domain.DocumentRepository
	llm  domain.LLMService
}

//go:generate moq -out moq_test.go -pkg usecase ../../domain DocumentRepository LLMService

func NewRagUsecase(repo domain.DocumentRepository, llm domain.LLMService) *RagUsecase {
	return &RagUsecase{
		repo: repo,
		llm:  llm,
	}
}

func (u *RagUsecase) RagAnswer(ctx context.Context, userQuestion string, searchLimit int16) error {
	// ==================================
	// 質問をベクトル（数値）に変換(Embedding)
	// ==================================
	vector, err := u.llm.GetEmbedding(userQuestion)
	if err != nil {
		return err
	}
	// ==================================
	// DBから意味が近い文章を検索(Retrieval)
	// ==================================
	retrievedContext, err := u.repo.SearchSimilar(ctx, vector, searchLimit)
	if err != nil {
		return err
	}
	// ==================================
	// LLMに「参考資料」と「質問」を渡して回答生成(Generation)
	// ==================================
	prompt := fmt.Sprintf(
		`あなたは優秀な日本人のアシスタントです。
		以下の【参考資料】に書かれている事実のみをもとに、ユーザーの【質問】に答えてください。
		重要：回答は【絶対に日本語】で出力してください。英語やその他の言語は使用しないでください。

		【参考資料】
		%s

		【質問】
		%s`,
		retrievedContext,
		userQuestion,
	)
	fmt.Print(prompt)
	// =================================
	// LLM回答
	// =================================
	response, err := u.llm.GenerateAnswer(prompt)
	if err != nil {
		return err
	}
	fmt.Println(response)
	return nil
}

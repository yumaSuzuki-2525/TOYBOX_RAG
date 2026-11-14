package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRagUsecase_RagAnswer(t *testing.T) {
	tests := []struct {
		// -----------------------------------------
		// テスト条件
		// -----------------------------------------
		name  string
		setup func(f *ragTestfields)
		// -----------------------------------------
		// 検証
		// -----------------------------------------
		assert  func(t *testing.T, f *ragTestfields)
		wantErr bool
	}{
		// ---------------------------------------------------
		// 正常系
		// ---------------------------------------------------
		{
			name: "正常系: 質問のベクトル化・類似検索・LLM回答生成が成功し、プロンプトに検索結果と質問が含まれる",
			assert: func(t *testing.T, f *ragTestfields) {
				calls := f.llm.GenerateAnswerCalls()
				if len(calls) != 1 {
					// 安全ガード兼バグ検知
					t.Fatalf("GenerateAnswer() の呼び出し回数 = %d回, 期待値 = 1回", len(calls))
				}
				gotPrompt := calls[0].Prompt
				if !strings.Contains(gotPrompt, "デフォルトのDBテキスト内容") {
					t.Errorf("プロンプトに取得した参考資料が含まれていません: %s", gotPrompt)
				}
			},
			wantErr: false,
		},
		// ---------------------------------------------------
		// 異常系
		// ---------------------------------------------------
		{
			name: "異常系: 質問のベクトル化(GetEmbedding)に失敗した場合、エラーを返す",
			setup: func(f *ragTestfields) {
				f.llm.GetEmbeddingFunc = func(text string) ([]float64, error) {
					return nil, errors.New("ベクトル変換エラー")
				}
			},
			wantErr: true,
		},
		{
			name: "異常系: 類似文章の検索(SearchSimilar)に失敗した場合、エラーを返す",
			setup: func(f *ragTestfields) {
				f.repo.SearchSimilarFunc = func(ctx context.Context, vector []float64, searchLimit int16) (string, error) {
					return "", errors.New("類似文章検索エラー")
				}
			},
			wantErr: true,
		},
		{
			name: "異常系: LLMの回答生成(GenerateAnswer)に失敗した場合、エラーを返す",
			setup: func(f *ragTestfields) {
				f.llm.GenerateAnswerFunc = func(prompt string) (string, error) {
					return "", errors.New("LLM回答生成エラー")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newDefaultRagFields()
			if tt.setup != nil {
				tt.setup(&f)
			}

			u := NewRagUsecase(f.repo, f.llm)
			if err := u.RagAnswer(context.Background(), "ユーザー質問", 10); (err != nil) != tt.wantErr {
				t.Errorf("RagAnswer() のエラー結果 = %v, エラー期待値 = %v", err, tt.wantErr)
			}

			if tt.assert != nil {
				tt.assert(t, &f)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// テスト用ヘルパー
// -----------------------------------------------------------------------------

type ragTestfields struct {
	repo *DocumentRepositoryMock
	llm  *LLMServiceMock
}

func newDefaultRagFields() ragTestfields {
	return ragTestfields{
		repo: &DocumentRepositoryMock{
			SearchSimilarFunc: func(ctx context.Context, vector []float64, searchLimit int16) (string, error) {
				return "デフォルトのDBテキスト内容", nil
			},
		},
		llm: &LLMServiceMock{
			GenerateAnswerFunc: func(prompt string) (string, error) {
				return "デフォルトのLLM回答", nil
			},
			GetEmbeddingFunc: func(text string) ([]float64, error) {
				return []float64{0.1, 0.2, 0.3}, nil
			},
		},
	}
}

package usecase

import (
	"context"
	"errors"
	"sample/internal/domain" // ※ご自身の環境に合わせてパスを調整してください
	"testing"
)

// ==========================================
// 1. 正常に動くモックAI
// ==========================================
type mockLLM struct{}

func (m *mockLLM) GetEmbedding(text string) ([]float64, error) {
	return []float64{0.1, 0.2, 0.3}, nil
}
func (m *mockLLM) GenerateAnswer(prompt string) (string, error) {
	return "これはAIのダミー回答です。", nil
}

// ==========================================
// 2. エラーを起こすモックAI
// ==========================================
type errorMockLLM struct{}

func (m *errorMockLLM) GetEmbedding(text string) ([]float64, error) {
	return nil, errors.New("AI通信に失敗しました")
}
func (m *errorMockLLM) GenerateAnswer(prompt string) (string, error) {
	return "", errors.New("回答生成に失敗しました")
}

// ==========================================
// 3. 正常に動くモックDB
// ==========================================
type mockRepo struct{}

func (m *mockRepo) SearchSimilar(ctx context.Context, vector []float64) (string, error) {
	return "これはDBから取得したダミーの参考資料です。", nil
}

// ==========================================
// 4. テスト本体（テーブル駆動テスト）
// ==========================================
func TestRagAnswer(t *testing.T) {
	// テストケースの表（テーブル）を定義
	tests := []struct {
		name     string            // テスト名
		question string            // 入力値
		mockLLM  domain.LLMService // 注入するモックAI
		wantErr  bool              // エラーを期待するか
		errMsg   string            // 期待するエラーメッセージ
	}{
		{
			name:     "正常系: 正しくAIの回答が生成される",
			question: "テストの質問",
			mockLLM:  &mockLLM{}, // 正常なAIをセット
			wantErr:  false,
		},
		{
			name:     "異常系: AIの通信でエラーが起きた場合は処理が中断される",
			question: "テストの質問",
			mockLLM:  &errorMockLLM{}, // 悪いAIをセット
			wantErr:  true,
			errMsg:   "AI通信に失敗しました",
		},
		// 今後「DBがエラーを返した場合」などもここに行を追加するだけでテスト可能です
	}

	// テーブルをループで回して実行
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックDBは共通で使う
			fakeRepo := &mockRepo{}
			// Usecaseにモックを注入
			ragUsecase := NewRagUsecase(fakeRepo, tt.mockLLM)

			// 実行
			err := ragUsecase.RagAnswer(context.Background(), tt.question)

			// 検証1: エラーの「有無」が期待通りか
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr = %v, but got err = %v", tt.wantErr, err)
				return // これ以上検証しない
			}

			// 検証2: エラーの「内容」が期待通りか
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

// 複数の参考資料を返すマルチモックDB
type multiMockRepo struct{}

func (m *multiMockRepo) SearchSimilar(ctx context.Context, vector []float64) (string, error) {
	// LIMIT 1 ではなく、複数の関連ドキュメントが結合されて返ってくる想定
	return "【資料1】Go言語のクリーンアーキテクチャについて\n【資料2】ポインタとインターフェースの使い分け", nil
}

func TestRagAnswer_MultipleContexts(t *testing.T) {
	fakeLLM := &mockLLM{}
	multiRepo := &multiMockRepo{}

	ragUsecase := NewRagUsecase(multiRepo, fakeLLM)

	// 実行
	err := ragUsecase.RagAnswer(context.Background(), "クリーンアーキテクチャについて教えて")

	if err != nil {
		t.Errorf("予期せぬエラー: %v", err)
	}
}

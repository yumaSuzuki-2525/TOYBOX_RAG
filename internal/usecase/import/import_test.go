package usecase

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestImportUsecase_ImportFromDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// -----------------------------------------
		// テスト条件
		// -----------------------------------------
		name           string
		files          []string
		overrideSrcDir string
		setup          func(f *importTestFields)
		// -----------------------------------------
		// 検証
		// -----------------------------------------
		assert  func(t *testing.T, f *importTestFields, srcDir, procDir string)
		wantErr bool
		// 呼び出し回数
		wantInsertCalls int
		wantMoveCalls   int
	}{
		// ---------------------------------------------------
		// 正常系
		// ---------------------------------------------------
		{
			name:  "正常系: サブディレクトリ内のMarkdownが2チャンクに分割され、正しい引数で保存・移動される",
			files: []string{"nested/dir/サンプル.md"},
			assert: func(t *testing.T, f *importTestFields, srcDir, procDir string) {
				insertCalls := f.repo.InsertDataCalls()
				if len(insertCalls) == 2 {
					if insertCalls[0].Chunk != "チャンク1" || insertCalls[0].VectorStr != "[0.1,0.2,0.3]" {
						t.Errorf("InsertData()[0] の引数 = %+v, 期待値 = Chunk: チャンク1, VectorStr: [0.1,0.2,0.3]", insertCalls[0])
					}
					if insertCalls[1].Chunk != "チャンク2" || insertCalls[1].VectorStr != "[0.1,0.2,0.3]" {
						t.Errorf("InsertData()[1] の引数 = %+v, 期待値 = Chunk: チャンク2, VectorStr: [0.1,0.2,0.3]", insertCalls[1])
					}
				}

				moveCalls := f.file.MoveToProcessedCalls()
				if len(moveCalls) == 1 {
					wantPath := filepath.Join(srcDir, "nested/dir/サンプル.md")
					if moveCalls[0].SourceDir != srcDir || moveCalls[0].ProcessedDir != procDir || moveCalls[0].CurrentPath != wantPath {
						t.Errorf("MoveToProcessed()[0] の引数 = %+v, 期待値 = SourceDir: %s, ProcessedDir: %s, CurrentPath: %s",
							moveCalls[0], srcDir, procDir, wantPath)
					}
				}
			},
			wantErr:         false,
			wantInsertCalls: 2,
			wantMoveCalls:   1,
		},
		{
			name:            "正常系: PDFファイルが読み込まれ、DB保存とファイル移動が行われる",
			files:           []string{"ドキュメント.pdf"},
			wantErr:         false,
			wantInsertCalls: 2,
			wantMoveCalls:   1,
		},
		{
			name:            "スキップ: 非対応拡張子(.png)は何も処理せず無視される",
			files:           []string{"画像.png"},
			wantErr:         false,
			wantInsertCalls: 0,
			wantMoveCalls:   0,
		},
		// ---------------------------------------------------
		// 異常系
		// ---------------------------------------------------
		{
			name:  "異常系: プレーンテキストの読み込み失敗時はスキップされ、DB保存も移動もされない",
			files: []string{"異常系.md"},
			setup: func(f *importTestFields) {
				f.file.ReadPlainFileFunc = func(path string) (string, error) {
					return "", errors.New("読み込みエラー")
				}
			},
			wantErr:         false,
			wantInsertCalls: 0,
			wantMoveCalls:   0,
		},
		{
			name:  "異常系: PDFの読み込み失敗時はスキップされ、DB保存も移動もされない",
			files: []string{"異常系.pdf"},
			setup: func(f *importTestFields) {
				f.file.ReadPdfFileFunc = func(path string) (string, error) {
					return "", errors.New("pdftotext実行エラー")
				}
			},
			wantErr:         false,
			wantInsertCalls: 0,
			wantMoveCalls:   0,
		},
		{
			name:  "異常系: チャンク分割エラー時はDB保存もファイル移動もされない",
			files: []string{"サンプル.md"},
			setup: func(f *importTestFields) {
				f.alz.ChunkTextBySemanticsFunc = func(text string, maxRunes int) ([]string, error) {
					return nil, errors.New("形態素解析エラー")
				}
			},
			wantErr:         false,
			wantInsertCalls: 0,
			wantMoveCalls:   0,
		},
		{
			name:  "境界値: 1チャンク目のベクトル化に失敗しても、2チャンク目は保存されファイル移動も行われる",
			files: []string{"サンプル.md"},
			setup: func(f *importTestFields) {
				callCount := 0
				f.imp.GetEmbeddingFunc = func(text string) ([]float64, error) {
					callCount++
					if callCount == 1 {
						return nil, errors.New("1チャンク目のみAPIエラー")
					}
					return []float64{0.1, 0.2, 0.3}, nil
				}
			},
			wantErr:         false,
			wantInsertCalls: 1,
			wantMoveCalls:   1,
		},
		{
			name:  "異常系: ファイル移動に失敗してもエラー終了せず処理を継続する",
			files: []string{"サンプル.md"},
			setup: func(f *importTestFields) {
				f.file.MoveToProcessedFunc = func(src, proc, current string) error {
					return errors.New("ファイル移動エラー")
				}
			},
			wantErr:         false,
			wantInsertCalls: 2,
			wantMoveCalls:   1,
		},
		{
			name:           "異常系: 存在しないディレクトリを指定した場合はエラーを返す",
			overrideSrcDir: filepath.Join(t.TempDir(), "not-exist"),
			wantErr:        true,
		},
		{
			name:  "異常系: DB保存に失敗したチャンクはスキップされ、次のチャンク処理が継続される",
			files: []string{"サンプル.md"},
			setup: func(f *importTestFields) {
				f.repo.InsertDataFunc = func(ctx context.Context, chunk string, vectorStr string) error {
					return errors.New("DB保存エラー")
				}
			},
			wantInsertCalls: 2, // 2チャンクとも InsertData 自体は呼ばれる
			wantMoveCalls:   1, // 現在の仕様ではループ後に移動処理が呼ばれる
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sourceDir, processedDir := setupTempDirs(t, tt.files)
			if tt.overrideSrcDir != "" {
				sourceDir = tt.overrideSrcDir
			}

			f := newDefaultImportFields()
			if tt.setup != nil {
				tt.setup(&f)
			}

			u := NewImportUsecase(f.repo, f.file, f.imp, f.alz)
			if err := u.ImportFromDirectory(context.Background(), sourceDir, processedDir); (err != nil) != tt.wantErr {
				t.Errorf("ImportFromDirectory() のエラー結果 = %v, エラー期待値 = %v", err, tt.wantErr)
			}

			if got := len(f.repo.InsertDataCalls()); got != tt.wantInsertCalls {
				t.Errorf("InsertData() の呼び出し回数 = %d回, 期待値 = %d回", got, tt.wantInsertCalls)
			}
			if got := len(f.file.MoveToProcessedCalls()); got != tt.wantMoveCalls {
				t.Errorf("MoveToProcessed() の呼び出し回数 = %d回, 期待値 = %d回", got, tt.wantMoveCalls)
			}
			if tt.assert != nil {
				tt.assert(t, &f, sourceDir, processedDir)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// テスト用ヘルパー
// -----------------------------------------------------------------------------

type importTestFields struct {
	repo *ImportRepositoryMock
	file *FileServiceMock
	imp  *ImportServiceMock
	alz  *AnalyzerServiceMock
}

func newDefaultImportFields() importTestFields {
	return importTestFields{
		file: &FileServiceMock{
			ReadPlainFileFunc: func(path string) (string, error) {
				return "デフォルトのテキスト内容", nil
			},
			ReadPdfFileFunc: func(path string) (string, error) {
				return "デフォルトのPDF内容", nil
			},
			MoveToProcessedFunc: func(src, proc, current string) error {
				return nil
			},
		},
		alz: &AnalyzerServiceMock{
			ChunkTextBySemanticsFunc: func(text string, maxRunes int) ([]string, error) {
				return []string{"チャンク1", "チャンク2"}, nil
			},
		},
		imp: &ImportServiceMock{
			GetEmbeddingFunc: func(text string) ([]float64, error) {
				return []float64{0.1, 0.2, 0.3}, nil
			},
		},
		repo: &ImportRepositoryMock{
			InsertDataFunc: func(ctx context.Context, chunk string, vectorStr string) error {
				return nil
			},
		},
	}
}

func setupTempDirs(t *testing.T, files []string) (string, string) {
	t.Helper()
	sourceDir := t.TempDir()
	processedDir := filepath.Join(t.TempDir(), "processed")

	for _, fileName := range files {
		fullPath := filepath.Join(sourceDir, fileName)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("dummy"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return sourceDir, processedDir
}

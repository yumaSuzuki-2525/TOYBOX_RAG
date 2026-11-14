package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// DocumentRepository は domain.DocumentRepository を満たすPostgres用の実装です
// dbを使用するためポイントレシーバにはdbを入れる。
type DocumentRepository struct {
	db *sql.DB
}

// NewDocumentRepository は DBコネクションを受け取ってリポジトリを作ります
func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) SearchSimilar(ctx context.Context, vector []float64, searchLimit int16) (string, error) {
	query := `
		SELECT content FROM documents ORDER BY embedding <-> $1 LIMIT $2;
		`
	// 検索時もJSONエンコードを使って確実にカンマ区切りの文字列にする
	vectorBytes, _ := json.Marshal(vector)
	vectorStr := string(vectorBytes)
	// 検索実行
	rows, err := r.db.Query(query, vectorStr, searchLimit)
	if err != nil {
		return "", fmt.Errorf("DB検索エラー: %v", err)
	}
	defer rows.Close()
	var contents []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return "", fmt.Errorf("行の読み込みエラー: %v", err)
		}
		contents = append(contents, content)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("行走査エラー: %v", err)
	}
	if len(contents) == 0 {
		return "", fmt.Errorf("該当する資料が見つかりませんでした")
	}
	return strings.Join(contents, "\n---\n"), nil
}

// DocumentRepository は domain.DocumentRepository を満たすPostgres用の実装です
// dbを使用するためポイントレシーバにはdbを入れる。
type ImportRepository struct {
	db *sql.DB
}

// NewDocumentRepository は DBコネクションを受け取ってリポジトリを作ります
func NewImportRepository(db *sql.DB) *ImportRepository {
	return &ImportRepository{db: db}
}

func (r *ImportRepository) InsertData(ctx context.Context, chunk string, vectorStr string) error {
	// DBへ保存
	query := `
		INSERT INTO documents 
		(content, embedding) VALUES ($1, $2)`
	_, err := r.db.Exec(query, chunk, vectorStr)
	if err != nil {
		return fmt.Errorf("DB登録エラー: %v", err)
	}
	return nil
}

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

func (r *DocumentRepository) SearchSimilar(ctx context.Context, vector []float64) (string, error) {
	query := `
		SELECT content FROM documents ORDER BY embedding <-> $1 LIMIT 1;
		`
	// 検索時もJSONエンコードを使って確実にカンマ区切りの文字列にする
	vectorBytes, _ := json.Marshal(vector)
	vectorStr := string(vectorBytes)
	// Goの配列形式[] をPostgresの配列形式{}ではなく、JSON風の[]のまま渡す(pgvectorの仕様)
	var retrievedContext string
	// 実際はDBにデータが入っていないとエラーになるため、ここではモックのテキストを代入します
	err := r.db.QueryRow(query, vectorStr).Scan(&retrievedContext)
	if err != nil {
		return "", fmt.Errorf("DB検索エラー: %v", err)
	}
	return retrievedContext, nil
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

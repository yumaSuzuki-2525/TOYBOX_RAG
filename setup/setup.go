package setup

import (
	"database/sql"
	"fmt"
	"sample/config"
)

type AppContainer struct {
	DB *sql.DB
}

// DB接続
func SetupApp(cfg *config.Config) (*sql.DB, error) {
	// Postgresqlに接続
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	return db, nil
}

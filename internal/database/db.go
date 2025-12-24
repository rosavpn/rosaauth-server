package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"rosaauth-server/internal/config"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type DB struct {
	Conn *sql.DB
}

func Connect(cfg *config.Config) (*DB, error) {
	conn, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	log.Info().Msg("Connected to database")
	return &DB{Conn: conn}, nil
}

func (db *DB) Migrate(migrationPath string) error {
	files, err := filepath.Glob(filepath.Join(migrationPath, "*.sql"))
	if err != nil {
		return err
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := db.Conn.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
		log.Info().Str("file", file).Msg("Migration executed")
	}
	return nil
}

func (db *DB) Close() error {
	return db.Conn.Close()
}

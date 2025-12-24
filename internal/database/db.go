package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
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
	root, err := os.OpenRoot(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to open migration directory: %w", err)
	}
	defer root.Close() //nolint:errcheck

	fsys := root.FS()
	files, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	for _, file := range files {
		content, err := fs.ReadFile(fsys, file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := db.Conn.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}
	return nil
}

func (db *DB) Close() error {
	return db.Conn.Close()
}

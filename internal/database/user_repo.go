package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"rosaauth-server/internal/models"

	"github.com/google/uuid"
)

type UserRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (r *UserRepo) CreateUser(ctx context.Context, u *models.User) error {
	query := `INSERT INTO users (id, email, password_hash, is_admin, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.ExecContext(ctx, query, u.ID, u.Email, u.PasswordHash, u.IsAdmin, u.CreatedAt)
	if err != nil {
		return fmt.Errorf("createUser: %w", err)
	}
	return nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, is_admin, created_at FROM users WHERE email = $1`
	row := r.DB.QueryRowContext(ctx, query, email)

	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getUserByEmail: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) CreateAdminIfNotExists(ctx context.Context, email, passwordHash string) error {
	u, err := r.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if u != nil {
		return nil
	}

	admin := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		IsAdmin:      true,
		CreatedAt:    time.Now(),
	}
	return r.CreateUser(ctx, admin)
}

func (r *UserRepo) ListUsers(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT u.id, u.email, u.is_admin, u.created_at, COUNT(tr.id) as record_count 
		FROM users u
		LEFT JOIN twofa_records tr ON u.id = tr.user_id
		GROUP BY u.id
		ORDER BY u.created_at DESC
	`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listUsers: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.IsAdmin, &u.CreatedAt, &u.RecordCount); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

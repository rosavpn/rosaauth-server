package database

import (
	"context"
	"database/sql"
	"fmt"

	"rosaauth-server/internal/models"

	"github.com/google/uuid"
)

type RecordRepo struct {
	DB *sql.DB
}

func NewRecordRepo(db *sql.DB) *RecordRepo {
	return &RecordRepo{DB: db}
}

func (r *RecordRepo) GetRecords(ctx context.Context, userID uuid.UUID) ([]models.TwoFARecordPayload, error) {
	query := `SELECT id, encrypted_data FROM twofa_records WHERE user_id = $1`
	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var records []models.TwoFARecordPayload
	for rows.Next() {
		var rec models.TwoFARecordPayload
		if err := rows.Scan(&rec.ID, &rec.EncryptedData); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *RecordRepo) ApplySyncOps(ctx context.Context, userID uuid.UUID, ops []models.SyncOperation) ([]models.TwoFARecordPayload, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	stmtUpsert := `
		INSERT INTO twofa_records (id, user_id, encrypted_data)
		VALUES ($1, $2, $3)
		ON CONFLICT (id, user_id) DO UPDATE 
		SET encrypted_data = EXCLUDED.encrypted_data
	`
	stmtDelete := `DELETE FROM twofa_records WHERE id = $1 AND user_id = $2`

	for _, op := range ops {
		switch op.Op {
		case models.SyncOpUpsert:
			dataStr := op.Data.EncryptedData
			if op.Data.ID == uuid.Nil {
				continue
			}
			if _, err := tx.ExecContext(ctx, stmtUpsert, op.Data.ID, userID, dataStr); err != nil {
				return nil, fmt.Errorf("upsert failed for %s: %w", op.Data.ID, err)
			}
		case models.SyncOpDelete:
			if op.Data.ID == uuid.Nil {
				continue
			}
			if _, err := tx.ExecContext(ctx, stmtDelete, op.Data.ID, userID); err != nil {
				return nil, fmt.Errorf("delete failed for %s: %w", op.Data.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetRecords(ctx, userID)
}

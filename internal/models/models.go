package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	RecordCount  int       `json:"record_count,omitempty"` // For Admin UI
}

type TwoFARecord struct {
	ID      uuid.UUID       `json:"id"`
	UserID  uuid.UUID       `json:"user_id"`
	Payload json.RawMessage `json:"payload"`
}

type SyncOp string

const (
	SyncOpUpsert SyncOp = "upsert"
	SyncOpDelete SyncOp = "delete"
)

type SyncOperation struct {
	Op   SyncOp             `json:"op"`
	Data TwoFARecordPayload `json:"data"`
}

type TwoFARecordPayload struct {
	ID      uuid.UUID       `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

// Admin API Models
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: refresh_tokens.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createRefreshToken = `-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
	$1,
	NOW(),
	NOW(),
	$2,
	$3
)
RETURNING token
`

type CreateRefreshTokenParams struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (string, error) {
	row := q.db.QueryRowContext(ctx, createRefreshToken, arg.Token, arg.UserID, arg.ExpiresAt)
	var token string
	err := row.Scan(&token)
	return token, err
}

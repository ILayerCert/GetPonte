// Package chat provides message persistence and retrieval for room conversations.
package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Message represents a single chat message.
type Message struct {
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"room_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	MsgType   string    `json:"msg_type"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository handles message persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new chat repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new message and returns the fully populated struct.
func (r *Repository) Create(ctx context.Context, roomID, userID uuid.UUID, content, msgType string) (*Message, error) {
	m := &Message{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO messages (room_id, user_id, content, msg_type)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, room_id, user_id, content, msg_type, created_at`,
		roomID, userID, content, msgType,
	).Scan(&m.ID, &m.RoomID, &m.UserID, &m.Content, &m.MsgType, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	// Attach the username for the response.
	err = r.pool.QueryRow(ctx, `SELECT username FROM users WHERE id = $1`, userID).Scan(&m.Username)
	if err != nil {
		return nil, fmt.Errorf("fetch username: %w", err)
	}

	return m, nil
}

// ListByRoom returns paginated messages for a room, ordered newest-first.
// If before is non-nil, only messages created before that timestamp are returned.
func (r *Repository) ListByRoom(ctx context.Context, roomID uuid.UUID, limit int, before *time.Time) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `SELECT m.id, m.room_id, m.user_id, u.username, m.content, m.msg_type, m.created_at
		 FROM messages m
		 INNER JOIN users u ON u.id = m.user_id
		 WHERE m.room_id = $1`
	args := []any{roomID}

	if before != nil {
		query += ` AND m.created_at < $2 ORDER BY m.created_at DESC LIMIT $3`
		args = append(args, *before, limit)
	} else {
		query += ` ORDER BY m.created_at DESC LIMIT $2`
		args = append(args, limit)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.UserID, &m.Username, &m.Content, &m.MsgType, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

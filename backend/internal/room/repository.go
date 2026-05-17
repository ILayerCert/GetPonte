// Package room provides the room domain: data access and HTTP handlers.
package room

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Room represents a collaboration room.
type Room struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedBy   uuid.UUID `json:"created_by"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
}

// Member represents a user's membership in a room.
type Member struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// RoomDetail combines a room with its member list.
type RoomDetail struct {
	Room    *Room    `json:"room"`
	Members []Member `json:"members"`
}

// Repository handles room persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new room repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new room and adds the creator as an admin member.
func (r *Repository) Create(ctx context.Context, name string, description *string, isPrivate bool, creatorID uuid.UUID) (*Room, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	room := &Room{}
	err = tx.QueryRow(ctx,
		`INSERT INTO rooms (name, description, created_by, is_private)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, description, created_by, is_private, created_at`,
		name, description, creatorID, isPrivate,
	).Scan(&room.ID, &room.Name, &room.Description, &room.CreatedBy, &room.IsPrivate, &room.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert room: %w", err)
	}

	// Creator automatically joins as admin.
	_, err = tx.Exec(ctx,
		`INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, 'admin')`,
		room.ID, creatorID,
	)
	if err != nil {
		return nil, fmt.Errorf("add creator as member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return room, nil
}

// ListByUser returns all rooms the given user is a member of.
func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]Room, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT r.id, r.name, r.description, r.created_by, r.is_private, r.created_at
		 FROM rooms r
		 INNER JOIN room_members rm ON rm.room_id = r.id
		 WHERE rm.user_id = $1
		 ORDER BY r.created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var rm Room
		if err := rows.Scan(&rm.ID, &rm.Name, &rm.Description, &rm.CreatedBy, &rm.IsPrivate, &rm.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		rooms = append(rooms, rm)
	}
	return rooms, rows.Err()
}

// FindByID retrieves a room by its primary key.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Room, error) {
	room := &Room{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, created_by, is_private, created_at
		 FROM rooms WHERE id = $1`, id,
	).Scan(&room.ID, &room.Name, &room.Description, &room.CreatedBy, &room.IsPrivate, &room.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find room: %w", err)
	}
	return room, nil
}

// GetMembers returns all members of a room.
func (r *Repository) GetMembers(ctx context.Context, roomID uuid.UUID) ([]Member, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.username, rm.role, rm.joined_at
		 FROM room_members rm
		 INNER JOIN users u ON u.id = rm.user_id
		 WHERE rm.room_id = $1
		 ORDER BY rm.joined_at`, roomID,
	)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.UserID, &m.Username, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// Join adds a user to a room. Returns false if already a member.
func (r *Repository) Join(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`INSERT INTO room_members (room_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		roomID, userID,
	)
	if err != nil {
		return false, fmt.Errorf("join room: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// Leave removes a user from a room. Returns false if not a member.
func (r *Repository) Leave(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`,
		roomID, userID,
	)
	if err != nil {
		return false, fmt.Errorf("leave room: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// IsMember checks whether a user belongs to a room.
func (r *Repository) IsMember(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`,
		roomID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}
	return exists, nil
}

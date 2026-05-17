package chat

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/interactionlayer/ponte/internal/auth"
	roomPkg "github.com/interactionlayer/ponte/internal/room"
)

// Handler exposes HTTP endpoints for message history.
type Handler struct {
	repo     *Repository
	roomRepo *roomPkg.Repository
}

// NewHandler creates a new chat handler.
func NewHandler(repo *Repository, roomRepo *roomPkg.Repository) *Handler {
	return &Handler{repo: repo, roomRepo: roomRepo}
}

// GetMessages handles GET /api/rooms/:id/messages?limit=50&before=<RFC3339>.
func (h *Handler) GetMessages(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}

	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	// Verify user is a member of the room.
	isMember, err := h.roomRepo.IsMember(c.Context(), roomID, claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if !isMember {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not a member of this room"})
	}

	// Parse optional query parameters.
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var before *time.Time
	if b := c.Query("before"); b != "" {
		t, err := time.Parse(time.RFC3339, b)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid 'before' timestamp, expected RFC3339"})
		}
		before = &t
	}

	messages, err := h.repo.ListByRoom(c.Context(), roomID, limit, before)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch messages"})
	}
	if messages == nil {
		messages = []Message{}
	}

	return c.JSON(messages)
}

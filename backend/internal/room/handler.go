package room

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/interactionlayer/ponte/internal/auth"
)

// Handler exposes HTTP endpoints for room management.
type Handler struct {
	repo *Repository
}

// NewHandler creates a new room handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// createRequest is the expected body for POST /api/rooms.
type createRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsPrivate   bool    `json:"is_private"`
}

// Create handles POST /api/rooms.
func (h *Handler) Create(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}

	var req createRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	room, err := h.repo.Create(c.Context(), req.Name, req.Description, req.IsPrivate, claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create room"})
	}

	return c.Status(fiber.StatusCreated).JSON(room)
}

// List handles GET /api/rooms.
func (h *Handler) List(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}

	rooms, err := h.repo.ListByUser(c.Context(), claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not list rooms"})
	}
	if rooms == nil {
		rooms = []Room{} // return empty array, not null
	}

	return c.JSON(rooms)
}

// Get handles GET /api/rooms/:id.
func (h *Handler) Get(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}

	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	room, err := h.repo.FindByID(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if room == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "room not found"})
	}

	members, err := h.repo.GetMembers(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch members"})
	}
	if members == nil {
		members = []Member{}
	}

	return c.JSON(RoomDetail{Room: room, Members: members})
}

// Join handles POST /api/rooms/:id/join.
func (h *Handler) Join(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}

	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	// Verify room exists.
	room, err := h.repo.FindByID(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if room == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "room not found"})
	}

	joined, err := h.repo.Join(c.Context(), roomID, claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not join room"})
	}
	if !joined {
		return c.JSON(fiber.Map{"message": "already a member"})
	}

	return c.JSON(fiber.Map{"message": "joined room"})
}

// Leave handles DELETE /api/rooms/:id/leave.
func (h *Handler) Leave(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}

	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	left, err := h.repo.Leave(c.Context(), roomID, claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not leave room"})
	}
	if !left {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "not a member of this room"})
	}

	return c.JSON(fiber.Map{"message": "left room"})
}

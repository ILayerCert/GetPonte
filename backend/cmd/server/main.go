// Ponte POC — Real-time collaboration platform backend.
// Entry point: loads config, connects to PostgreSQL, runs migrations,
// sets up routes and the WebSocket hub, then starts the HTTP server.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/interactionlayer/ponte/internal/auth"
	"github.com/interactionlayer/ponte/internal/chat"
	"github.com/interactionlayer/ponte/internal/config"
	"github.com/interactionlayer/ponte/internal/db"
	"github.com/interactionlayer/ponte/internal/room"
	"github.com/interactionlayer/ponte/internal/user"
	"github.com/interactionlayer/ponte/internal/ws"
)

func main() {
	// ── Configuration ───────────────────────────────────────────────
	cfg := config.Load()

	// ── Database ────────────────────────────────────────────────────
	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()

	// Run migrations from the ./migrations directory.
	if err := db.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	// ── Repositories ────────────────────────────────────────────────
	userRepo := user.NewRepository(pool)
	roomRepo := room.NewRepository(pool)
	chatRepo := chat.NewRepository(pool)

	// ── Handlers ────────────────────────────────────────────────────
	userHandler := user.NewHandler(userRepo, cfg.JWTSecret)
	roomHandler := room.NewHandler(roomRepo)
	chatHandler := chat.NewHandler(chatRepo, roomRepo)
	hub := ws.NewHub(chatRepo, roomRepo, cfg.JWTSecret)

	// ── Fiber App ───────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "Ponte POC",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware.
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	// ── Routes ──────────────────────────────────────────────────────

	// Health check.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Auth (public).
	authGroup := app.Group("/api/auth")
	authGroup.Post("/register", userHandler.Register)
	authGroup.Post("/login", userHandler.Login)

	// Authenticated API.
	api := app.Group("/api", auth.Middleware(cfg.JWTSecret))

	api.Get("/users/me", userHandler.Me)

	api.Post("/rooms", roomHandler.Create)
	api.Get("/rooms", roomHandler.List)
	api.Get("/rooms/:id", roomHandler.Get)
	api.Post("/rooms/:id/join", roomHandler.Join)
	api.Delete("/rooms/:id/leave", roomHandler.Leave)

	api.Get("/rooms/:id/messages", chatHandler.GetMessages)

	// WebSocket (uses query-param auth, not header middleware).
	app.Use("/ws", hub.UpgradeMiddleware())
	app.Get("/ws", hub.Handler())

	// ── Start Server ────────────────────────────────────────────────
	go func() {
		addr := ":" + cfg.Port
		log.Printf("Ponte server starting on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

// customErrorHandler provides consistent JSON error responses for unhandled errors.
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

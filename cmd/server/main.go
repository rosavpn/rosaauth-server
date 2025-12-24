package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rosaauth-server/internal/config"
	"rosaauth-server/internal/database"
	"rosaauth-server/internal/handlers"
	"rosaauth-server/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. Config & Logger
	cfg := config.LoadConfig()
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil || cfg.LogLevel == "disabled" {
		if cfg.LogLevel == "disabled" {
			level = zerolog.Disabled
		} else {
			level = zerolog.InfoLevel // Default
		}
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339

	// 2. Database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to database")
	}
	defer db.Close() //nolint:errcheck

	if err := db.Migrate("internal/database/migrations"); err != nil {
		log.Fatal().Err(err).Msg("Migration failed")
	}

	// 3. Repositories
	userRepo := database.NewUserRepo(db.Conn)
	recordRepo := database.NewRecordRepo(db.Conn)

	// 4. Initial Admin User
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to hash admin password")
	}
	if err := userRepo.CreateAdminIfNotExists(context.Background(), cfg.AdminEmail, string(hash)); err != nil {
		log.Fatal().Err(err).Msg("Failed to ensure admin user exists")
	}

	// 5. Handlers & Middleware
	authMw := middleware.NewAuthMiddleware(cfg)
	authHandler := handlers.NewAuthHandler(userRepo, authMw)
	syncHandler := handlers.NewSyncHandler(recordRepo, userRepo)
	adminHandler := handlers.NewAdminHandler(userRepo)

	// 6. Fiber App
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	if level != zerolog.Disabled {
		app.Use(logger.New(logger.Config{
			Format:     `{"level":"info","time":"${time}","status":${status},"method":"${method}","path":"${path}","error":"${error}"}` + "\n",
			TimeFormat: time.RFC3339,
			Output:     os.Stderr,
		}))
	}
	app.Use(cors.New())

	// Global Rate Limiter
	app.Use(limiter.New(limiter.Config{
		Max:        cfg.GlobalRateLimit,
		Expiration: 1 * time.Minute,
	}))

	// Login Rate Limiter
	loginLimiter := limiter.New(limiter.Config{
		Max:        cfg.LoginRateLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Limit by IP
		},
	})

	// Static
	app.Static("/", "./web")

	// API Routes
	api := app.Group("/api/v1")
	api.Post("/login", loginLimiter, authHandler.Login)
	api.Post("/sync", authMw.Protect(), syncHandler.Sync)

	// Admin Routes
	admin := app.Group("/admin")
	// Middleware to protect admin routes
	admin.Use(authMw.Protect())
	admin.Use(authMw.AdminOnly())

	admin.Get("/users", adminHandler.ListUsers)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Delete("/users/:id", adminHandler.DeleteUser)

	// 7. Start
	go func() {
		log.Info().Str("port", cfg.Port).Msg("Server starting")
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// 8. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}
	log.Info().Msg("Server exited")
}

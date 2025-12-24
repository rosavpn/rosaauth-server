package handlers

import (
	"rosaauth-server/internal/database"
	"rosaauth-server/internal/middleware"
	"rosaauth-server/internal/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	UserRepo *database.UserRepo
	AuthMw   *middleware.AuthMiddleware
}

func NewAuthHandler(userRepo *database.UserRepo, authMw *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		UserRepo: userRepo,
		AuthMw:   authMw,
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user, err := h.UserRepo.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, err := h.AuthMw.GenerateToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(models.LoginResponse{Token: token})
}

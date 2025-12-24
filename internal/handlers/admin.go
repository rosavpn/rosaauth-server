package handlers

import (
	"rosaauth-server/internal/database"
	"rosaauth-server/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	UserRepo *database.UserRepo
}

func NewAdminHandler(userRepo *database.UserRepo) *AdminHandler {
	return &AdminHandler{UserRepo: userRepo}
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.UserRepo.ListUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch users"})
	}
	return c.JSON(users)
}

func (h *AdminHandler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	existing, err := h.UserRepo.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB error"})
	}
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User already exists"})
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		IsAdmin:      req.IsAdmin,
		CreatedAt:    time.Now(),
	}

	if err := h.UserRepo.CreateUser(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	if err := h.UserRepo.DeleteUser(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete user"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

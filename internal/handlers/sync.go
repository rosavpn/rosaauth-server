package handlers

import (
	"rosaauth-server/internal/database"
	"rosaauth-server/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type SyncHandler struct {
	RecordRepo *database.RecordRepo
	UserRepo   *database.UserRepo
}

func NewSyncHandler(recordRepo *database.RecordRepo, userRepo *database.UserRepo) *SyncHandler {
	return &SyncHandler{
		RecordRepo: recordRepo,
		UserRepo:   userRepo,
	}
}

func (h *SyncHandler) Sync(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID in token"})
	}

	var ops []models.SyncOperation
	if err := c.BodyParser(&ops); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	newRecords, err := h.RecordRepo.ApplySyncOps(c.Context(), userID, ops)
	if err != nil {
		log.Error().Err(err).Msg("Failed to apply sync ops")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Sync failed"})
	}

	return c.JSON(newRecords)
}

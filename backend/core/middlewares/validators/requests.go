package validators

import (
	"backend/core/repositories"
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func ValidatePostMatchRequest(c fiber.Ctx) error {
	var body struct {
		ToUserId int `json:"to_user_id"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	c.Locals("ToUserId", body.ToUserId)
	return c.Next()
}

func ValidateMatchOwnership(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	matchID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid match request ID",
		})
	}

	toUserID, err := repositories.SelectToUserIdFromMatchRequests(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Match request not found",
		})
	}

	if toUserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not authorized to update this request",
		})
	}

	c.Locals("matchID", matchID)
	return c.Next()
}

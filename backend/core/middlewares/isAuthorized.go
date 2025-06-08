package middlewares

import (
	"backend/core/services"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// used to check whether the user is authorized
func IsAuthorized(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is missing",
		})
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is invalid",
		})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	userUUID, err := services.ExtractUUID(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	c.Locals("userUUID", userUUID)

	return c.Next()
}

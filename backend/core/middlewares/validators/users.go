package validators

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func ValidateLanguages(c fiber.Ctx) error {
	var body struct {
		Native []int `json:"native"`
		Target []int `json:"target"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	if len(body.Native) == 0 && len(body.Target) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one of 'native' or 'target' must be provided",
		})
	}

	c.Locals("languages", body)
	return c.Next()
}

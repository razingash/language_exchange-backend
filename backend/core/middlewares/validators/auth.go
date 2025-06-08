package validators

import (
	"encoding/json"
	"regexp"

	"github.com/gofiber/fiber/v3"
)

func ValidateRegisterInfo(c fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"full_name"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	if len(body.Username) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"username": "Username must be at least 6 characters long",
		})
	}

	if len(body.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"password": "Password must be at least 6 characters long",
		})
	}

	if !isValidEmail(body.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"email": "Invalid email format",
		})
	}

	c.Locals("body", body)

	return c.Next()
}

func ValidateLoginInfo(c fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	if len(body.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"password": "Password must be at least 6 characters long",
		})
	}

	if !isValidEmail(body.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"email": "Invalid email format",
		})
	}

	c.Locals("body", body)

	return c.Next()
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

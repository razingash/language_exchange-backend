package handlers

import (
	"backend/core/repositories"
	"backend/core/services"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func Register(c fiber.Ctx) error {
	body := c.Locals("body").(struct {
		Username string `json:"full_name"`
		Password string `json:"password"`
		Email    string `json:"email"`
	})

	user, err := services.RegisterUser(body.Username, body.Password, body.Email)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "User with this username already exists",
			})
		}
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error during user registration",
		})
	}
	if err := c.JSON(&body); err != nil {
		return err
	}

	accessToken := services.GenerateAccessToken(user.ID)
	refreshToken := services.GenerateRefreshToken(user.ID)

	err = repositories.SaveAccessToken(user.ID, accessToken)
	if err != nil {
		return nil
	}

	err = repositories.SaveRefreshToken(user.ID, refreshToken)
	if err != nil {
		return nil
	}

	return c.JSON(fiber.Map{
		"access":  accessToken,
		"refresh": refreshToken,
	})
}

func Login(c fiber.Ctx) error {
	body := c.Locals("body").(struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	})

	user, username, err := services.LoginUser(body.Email, body.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Incorrect login or password",
		})
	}

	accessToken := services.GenerateAccessToken(user.ID)
	refreshToken := services.GenerateRefreshToken(user.ID)

	err = repositories.SaveAccessToken(user.ID, accessToken)
	if err != nil {
		return nil
	}

	err = repositories.SaveRefreshToken(user.ID, refreshToken)
	if err != nil {
		return nil
	}

	return c.JSON(fiber.Map{
		"access":    accessToken,
		"refresh":   refreshToken,
		"id":        user.ID,
		"email":     body.Email,
		"full_name": username,
	})
}

// validates tokens, both refresh and access
func ValidateToken(c fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	isTokenValid := services.ValidateToken(body.Token)

	return c.JSON(fiber.Map{
		"isValid": isTokenValid,
	})
}

func RefreshAccessToken(c fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	errCode, newAccessToken := services.GetNewAccessToken(body.Token)

	if errCode == 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	return c.JSON(fiber.Map{
		"access": newAccessToken,
	})
}

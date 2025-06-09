package handlers

import (
	"backend/core/repositories"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

func GetUserInfo(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	user, err := repositories.SelectUserInfo(userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query user",
		})
	}

	rows, err := repositories.SelectUserLanguages(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query user languages",
		})
	}
	defer rows.Close()

	var languages []map[string]interface{}

	for rows.Next() {
		var languageID int
		var langType string
		if err := rows.Scan(&languageID, &langType); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan user language",
			})
		}
		languages = append(languages, map[string]interface{}{
			"language_id": languageID,
			"type":        langType,
		})
	}

	return c.JSON(fiber.Map{
		"id":        user.ID,
		"email":     user.Email,
		"full_name": user.FullName,
		"languages": languages,
	})
}

func GetTargetedUsers(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	native := c.Query("native")
	target := c.Query("target")

	rows, err := repositories.SelectTargetedUsers(native, target, userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}
	defer rows.Close()

	var users []fiber.Map

	for rows.Next() {
		var id int
		var email, fullName string
		if err := rows.Scan(&id, &email, &fullName); err != nil {
			continue
		}
		users = append(users, fiber.Map{
			"id":        id,
			"email":     email,
			"full_name": fullName,
		})
	}

	return c.JSON(users)
}

func UpdateUserLanguages(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	err := repositories.UpdateSelectedLanguages(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update users",
		})
	}

	user, err := repositories.SelectUserInfo(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	rows, err := repositories.SelectUserLanguages(userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user languages",
		})
	}

	defer rows.Close()

	var languages []fiber.Map
	for rows.Next() {
		var langID int
		var langType string
		if err := rows.Scan(&langID, &langType); err != nil {
			continue
		}
		languages = append(languages, fiber.Map{
			"language_id": langID,
			"type":        langType,
		})
	}

	return c.JSON(fiber.Map{
		"id":        user.ID,
		"email":     user.Email,
		"full_name": user.FullName,
		"languages": languages,
	})
}

package handlers

import (
	"backend/core/repositories"
	"database/sql"
	"errors"
	"log"

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
		log.Println(err)
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

	rows, err := repositories.SelectTargetedUsers(target, native, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}
	defer rows.Close()

	type userData struct {
		ID       int
		Email    string
		FullName string
		Natives  []int
		Targets  []int
	}

	usersMap := make(map[int]*userData)

	for rows.Next() {
		var (
			id       int
			email    string
			fullName string
			langID   sql.NullInt32
			langType sql.NullString
		)

		if err := rows.Scan(&id, &email, &fullName, &langID, &langType); err != nil {
			continue
		}

		user, exists := usersMap[id]
		if !exists {
			user = &userData{
				ID:       id,
				Email:    email,
				FullName: fullName,
			}
			usersMap[id] = user
		}

		if !langID.Valid || !langType.Valid {
			continue
		}

		switch langType.String {
		case "native":
			user.Natives = append(user.Natives, int(langID.Int32))
		case "target":
			user.Targets = append(user.Targets, int(langID.Int32))
		}
	}

	var users []fiber.Map
	for _, u := range usersMap {
		users = append(users, fiber.Map{
			"id":        u.ID,
			"email":     u.Email,
			"full_name": u.FullName,
			"native":    u.Natives,
			"target":    u.Targets,
		})
	}

	return c.JSON(users)
}

func UpdateUserLanguages(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	langs := c.Locals("languages").(repositories.Languages)

	err := repositories.UpdateSelectedLanguages(userID, langs)
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

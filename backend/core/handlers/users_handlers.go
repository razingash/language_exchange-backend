package handlers

import (
	"backend/core/db"
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

func GetUserInfo(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var user struct {
		ID       int    `json:"id"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}

	err := db.DB.QueryRow(context.Background(), `
		SELECT id, email, full_name
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.FullName)

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

	rows, err := db.DB.Query(context.Background(), `
		SELECT language_id, type
		FROM user_languages
		WHERE user_id = $1
	`, userID)
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

	ctx := context.Background()
	var rows pgx.Rows
	var err error

	if native != "" && target != "" {
		rows, err = db.DB.Query(ctx, `
			SELECT u.id, u.email, u.full_name
			FROM users u
			JOIN user_languages ul1 ON ul1.user_id = u.id AND ul1.type = 'native' AND ul1.language_id = $1
			JOIN user_languages ul2 ON ul2.user_id = u.id AND ul2.type = 'target' AND ul2.language_id = $2
			WHERE u.id != $3
		`, target, native, userID)
	} else {
		rows, err = db.DB.Query(ctx, `
			SELECT id, email, full_name FROM users WHERE id != $1
		`, userID)
	}

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

	langs, ok := c.Locals("languages").(struct {
		Native []int `json:"native"`
		Target []int `json:"target"`
	})
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid language data",
		})
	}

	ctx := context.Background()

	for _, langID := range langs.Native {
		_, _ = db.DB.Exec(ctx, `
			INSERT INTO user_languages (user_id, language_id, type)
			VALUES ($1, $2, 'native')
			ON CONFLICT (user_id, language_id, type) DO NOTHING
		`, userID, langID)
	}
	for _, langID := range langs.Target {
		_, _ = db.DB.Exec(ctx, `
			INSERT INTO user_languages (user_id, language_id, type)
			VALUES ($1, $2, 'target')
			ON CONFLICT (user_id, language_id, type) DO NOTHING
		`, userID, langID)
	}

	var user struct {
		ID       int    `json:"id"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}

	err := db.DB.QueryRow(ctx, `
		SELECT id, email, full_name FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.FullName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	rows, err := db.DB.Query(ctx, `
		SELECT language_id, type FROM user_languages WHERE user_id = $1
	`, userID)

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

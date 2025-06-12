package handlers

import (
	"backend/core/repositories"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

func CreateMatchRequest(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	toUserID := c.Locals("ToUserId").(int)
	value, err := strconv.Atoi(userID)

	if err != nil {
		log.Println(1, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "broken jwt token", // ip should be banned, bcs token is forged
		})
	}

	if value == toUserID {
		log.Println(2, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot send match request to yourself",
		})
	}
	requestID, err := repositories.InsertMatchRequest(userID, toUserID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Match request already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create match request",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Match request sent",
		"request_id": requestID,
		"from_user":  userID,
		"to_user":    toUserID,
		"status":     "pending",
	})
}

func GetIncomingMatchRequest(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	rows, err := repositories.SelectIncomingMatchRequests(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch incoming match requests",
		})
	}

	var requests []map[string]interface{}
	for rows.Next() {
		var id, fromUserID int
		var status, fullName string
		var createdAt time.Time

		if err := rows.Scan(&id, &fromUserID, &status, &createdAt, &fullName); err != nil {
			continue
		}
		requests = append(requests, fiber.Map{
			"id":           id,
			"from_user_id": fromUserID,
			"status":       status,
			"created_at":   createdAt,
			"full_name":    fullName,
		})
	}

	return c.JSON(fiber.Map{
		"incoming_requests": requests,
	})
}

func GetOutgoingMatchRequest(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	rows, err := repositories.SelectOutcomingMatchRequests(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch outgoing match requests",
		})
	}

	var requests []map[string]interface{}
	for rows.Next() {
		var id, toUserID int
		var status, fullName string
		var createdAt time.Time

		if err := rows.Scan(&id, &toUserID, &status, &createdAt, &fullName); err != nil {
			continue
		}
		requests = append(requests, fiber.Map{
			"id":         id,
			"to_user_id": toUserID,
			"status":     status,
			"created_at": createdAt,
			"full_name":  fullName,
		})
	}

	return c.JSON(fiber.Map{
		"outgoing_requests": requests,
	})
}

func GetAcceptedMatchRequest(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	rows, err := repositories.SelectAcceptedMatchRequests(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch accepted match requests",
		})
	}

	var matches []map[string]interface{}
	for rows.Next() {
		var id, fromID, toID int
		var status, fullName string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &fromID, &toID, &status, &createdAt, &updatedAt, &fullName); err != nil {
			continue
		}

		matches = append(matches, fiber.Map{
			"id":           id,
			"from_user_id": fromID,
			"to_user_id":   toID,
			"status":       status,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
			"full_name":    fullName,
		})
	}

	return c.JSON(fiber.Map{
		"accepted_matches": matches,
	})
}

func PutAcceptMatchRequest(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	err := repositories.ChangeMatchRequestStatusToAccepted(userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to accept match request",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Match request accepted",
	})
}

func PutDeclineMatchRequest(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	err := repositories.ChangeMatchRequestStatusToDeclined(userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decline match request",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Match request declined",
	})
}

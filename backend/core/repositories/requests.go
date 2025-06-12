package repositories

import (
	"backend/core/db"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func InsertMatchRequest(userID string, toUserID int) (int, error) {
	var requestID int
	err := db.DB.QueryRow(context.Background(), `
	INSERT INTO match_requests (from_user_id, to_user_id)
		VALUES ($1, $2)
		ON CONFLICT (from_user_id, to_user_id) DO NOTHING
		RETURNING id
	`, userID, toUserID).Scan(&requestID)

	return requestID, err
}

func SelectToUserIdFromMatchRequests(matchID string) (string, error) {
	var toUserID string
	err := db.DB.QueryRow(context.Background(), `
		SELECT to_user_id FROM match_requests WHERE id = $1
	`, matchID).Scan(&toUserID)

	return toUserID, err
}

func SelectOutcomingMatchRequests(userID string) (pgx.Rows, error) {
	rows, err := db.DB.Query(context.Background(), `
		SELECT 
			mr.id, mr.to_user_id, mr.status, mr.created_at, u.full_name
		FROM match_requests mr
		JOIN users u ON mr.to_user_id = u.id
		WHERE mr.from_user_id = $1 AND mr.status = 'pending'
	`, userID)

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func SelectIncomingMatchRequests(userID string) (pgx.Rows, error) {
	rows, err := db.DB.Query(context.Background(), `
		SELECT 
			mr.id, mr.from_user_id, mr.status, mr.created_at, u.full_name
		FROM match_requests mr
		JOIN users u ON mr.from_user_id = u.id
		WHERE mr.to_user_id = $1 AND mr.status = 'pending'
	`, userID)

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func SelectAcceptedMatchRequests(userID string) (pgx.Rows, error) {
	rows, err := db.DB.Query(context.Background(), `
		SELECT 
			mr.id, mr.from_user_id, mr.to_user_id, mr.status, mr.created_at, mr.updated_at,
			CASE 
				WHEN mr.from_user_id = $1 THEN u2.full_name
				ELSE u1.full_name
			END AS full_name
		FROM match_requests mr
		JOIN users u1 ON u1.id = mr.from_user_id
		JOIN users u2 ON u2.id = mr.to_user_id
		WHERE mr.status = 'accepted' AND (mr.from_user_id = $1 OR mr.to_user_id = $1)
	`, userID)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func ChangeMatchRequestStatusToDeclined(userID string) error {
	_, err := db.DB.Exec(context.Background(), `
		UPDATE match_requests
		SET status = 'declined', updated_at = $1
		WHERE id = $2
	`, time.Now(), userID)

	return err
}

func ChangeMatchRequestStatusToAccepted(userID string) error {
	_, err := db.DB.Exec(context.Background(), `
		UPDATE match_requests
		SET status = 'accepted', updated_at = $1
		WHERE id = $2
	`, time.Now(), userID)

	return err
}

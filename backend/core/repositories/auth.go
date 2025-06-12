package repositories

import (
	"backend/core/db"
	"context"
	"time"
)

type User struct {
	ID       string
	Password string
}

func CreateUser(username string, password string, email string) error {
	_, err := db.DB.Exec(context.Background(), `
        INSERT INTO users (full_name, password_hash, email) 
        VALUES ($1, $2, $3) 
        RETURNING id`,
		username, password, email,
	)

	if err != nil {
		return err
	}
	return nil
}

func SaveAccessToken(uuid string, accessToken string) error {
	_, err := db.DB.Exec(context.Background(), `
		INSERT INTO access_tokens (user_id, token, expires_at, created_at) 
		VALUES ($1, $2, $3, $4)`,
		uuid, accessToken, time.Now().Add(15*time.Minute), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func SaveRefreshToken(uuid string, refreshToken string) error {
	_, err := db.DB.Exec(context.Background(), `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, revoked) 
		VALUES ($1, $2, $3, $4, false)`,
		uuid, refreshToken, time.Now().Add(24*time.Hour), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func LoginUser(email string) (string, string, string, error) {
	var userID string
	var userPassword string
	var username string

	err := db.DB.QueryRow(context.Background(), `
		SELECT id, password_hash, full_name
		FROM users 
		WHERE email = $1
	`, email).Scan(&userID, &userPassword, &username)

	if err != nil {
		return "", "", "", err
	}
	return userID, userPassword, username, nil
}

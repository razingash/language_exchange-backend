package services

import (
	"backend/core/db"
	"backend/main/config"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Payload struct {
	Sub string `json:"sub"` // id
	Exp int64  `json:"exp"` // expiration time
	Iat int64  `json:"iat"` // issued at
}

func GenerateAccessToken(userID string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	expiry := int64(600) // 10 minutes

	payloadStruct := Payload{
		Sub: userID,
		Exp: time.Now().Unix() + expiry,
		Iat: time.Now().Unix(),
	}
	payloadBytes, _ := json.Marshal(payloadStruct)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signature := signHMAC(header+"."+payload, config.JWT_SECRET)

	token := header + "." + payload + "." + signature
	return token
}

func GenerateRefreshToken(userID string) string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)

	h := hmac.New(sha256.New, []byte(config.JWT_SECRET))
	h.Write([]byte(userID))
	h.Write(randomBytes)

	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// gets id from access token payload, validated token must be passed
func ExtractID(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		if ValidateAccessToken(parts) {
			payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				return "", fmt.Errorf("failed to decode payload")
			}

			var payloadData Payload
			err = json.Unmarshal(payloadBytes, &payloadData)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal payload")
			}

			return payloadData.Sub, nil
		}
	}
	return "", nil
}

func ValidateToken(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		return ValidateAccessToken(parts)
	} else {
		return ValidateRefreshToken(token)
	}
}

func ValidateRefreshToken(token string) bool {
	var expTime time.Time
	var revoked bool

	err := db.DB.QueryRow(context.Background(), `
		SELECT expires_at, revoked FROM refresh_tokens WHERE token = $1
	`, token).Scan(&expTime, &revoked)

	if err != nil || revoked || time.Now().After(expTime) {
		return false
	}

	return true
}

func ValidateAccessToken(token_parts []string) bool {
	header, payload, receivedSig := token_parts[0], token_parts[1], token_parts[2]
	expectedSig := signHMAC(header+"."+payload, config.JWT_SECRET)

	if receivedSig != expectedSig {
		return false //, "Invalid signature"
	}

	return CheckAccessTokenRelevance(payload)
}

func CheckAccessTokenRelevance(payload string) bool {
	payloadBytes, _ := base64.RawURLEncoding.DecodeString(payload)
	var payloadData Payload
	json.Unmarshal(payloadBytes, &payloadData)

	if time.Now().Unix() > payloadData.Exp {
		return false // "Token expired"
	}

	return true // "Valid token"
}

// returns a new access token if the recharge token is OK, otherwise an error
func GetNewAccessToken(token string) (int, string) {
	var userID string
	var expTime time.Time
	var revoked bool

	err := db.DB.QueryRow(context.Background(), `
		SELECT user_id, expires_at, revoked 
		FROM refresh_tokens 
		WHERE token = $1
	`, token).Scan(&userID, &expTime, &revoked)

	if err != nil || revoked || time.Now().After(expTime) {
		return 1, ""
	}

	return 0, GenerateAccessToken(userID)
}

// signature generation
func signHMAC(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(signature)
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

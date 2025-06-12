package services

import (
	"backend/core/repositories"
	"fmt"
)

func RegisterUser(username string, password string, email string) error {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	err = repositories.CreateUser(username, hashedPassword, email)

	if err != nil {
		return err
	}

	return nil
}

func LoginUser(email string, password string) (*repositories.User, string, error) {
	userID, userPassword, username, err := repositories.LoginUser(email)

	if err != nil {
		return nil, "", err
	}

	if !CheckPassword(password, userPassword) {
		return nil, "", fmt.Errorf("invalid password")
	}

	return &repositories.User{ID: userID}, username, nil
}

package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	JWT_SECRET    string
	Database_Url  string
	Database_Name string
	PORT          string
	IsInDocker    bool = false
)

// function to set environment variables
func LoadConfig() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Application startup via docker") // most likely
		IsInDocker = true
	}

	JWT_SECRET = os.Getenv("JWT_SECRET")
	Database_Name = os.Getenv("DB_NAME")
	PORT = os.Getenv("PORT")
	var Database_User = os.Getenv("DB_USER")
	var Database_Password = os.Getenv("DB_PASSWORD")
	var Database_Host = os.Getenv("DB_HOST")
	var Database_Port = os.Getenv("DB_PORT")

	if JWT_SECRET == "" {
		log.Fatal("JWT_SECRET not setted in the environment")
	}
	if Database_Name == "" {
		log.Fatal("Database_Name not setted in the environment")
	}
	if Database_User == "" {
		log.Fatal("Database_User not setted in the environment")
	}
	if Database_Password == "" {
		log.Fatal("Database_Password not setted in the environment")
	}
	if Database_Host == "" {
		log.Fatal("Database_Host not setted in the environment")
	}
	if Database_Port == "" {
		log.Fatal("Database_Port not setted in the environment")
	}

	Database_Url = fmt.Sprintf(
		"postgresql://%v:%v@%v:%v/%v?sslmode=disable",
		Database_User, Database_Password, Database_Host, Database_Port, Database_Name,
	)
}

package main

import (
	"log"

	"os"

	"github.com/Arinji2/sense-backend/api"
	"github.com/joho/godotenv"
)

var pbLink = "https://db-word.arinji.com"

func PocketbaseAdminLogin() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")

	}

	identityEmail := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")

	if identityEmail == "" || password == "" {
		log.Fatal("Environment Variables not present to authenticate Admin")

	}

	body := map[string]string{
		"identity": identityEmail,
		"password": password,
	}

	client := api.NewApiClient(pbLink)
	result, err := client.SendRequest("POST", "/api/admins/auth-with-password", body)

	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	token, ok := result["token"].(string)
	if !ok || token == "" {
		log.Fatalf("Token not found or not a string")
	}

	return token

}

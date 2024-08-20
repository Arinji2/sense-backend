package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var pbLink = "https://db-word.arinji.com"

func PocketbaseAdminLogin() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")

	}

	identityEmail := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")

	if identityEmail == "" || password == "" {
		log.Fatal("Environment Variables not present to authenticate Admin")

	}
	fetchString := fmt.Sprintf("%s/%s", pbLink, "api/admins/auth-with-password")
	bodyParams := map[string]string{
		"identity": identityEmail,
		"password": password,
	}

	jsonData, err := json.Marshal(bodyParams)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", fetchString, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	fmt.Println(result)

}

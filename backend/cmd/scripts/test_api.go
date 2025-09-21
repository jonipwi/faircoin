package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	// First, login to get a token
	loginData := map[string]string{
		"username": "alice123",
		"password": "password123",
	}

	loginJSON, _ := json.Marshal(loginData)
	resp, err := http.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResponse struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		log.Fatalf("Failed to decode login response: %v", err)
	}

	fmt.Printf("Login successful, token: %s...\n", loginResponse.Token[:20])

	// Now test the transaction volume endpoint
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/admin/transaction-volume", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+loginResponse.Token)

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Transaction volume request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Transaction volume API response (Status: %d):\n%s\n", resp.StatusCode, string(body))
}

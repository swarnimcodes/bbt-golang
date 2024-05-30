package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/swarnimcodes/bitespeed-backend-task/utils"
)

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func GenerateJWT(w http.ResponseWriter, r *http.Request) {

	var creds UserCredentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		message := "Invalid JSON input"
		utils.SendErrorResponse(w, message, http.StatusBadRequest)
		return
	}

	// TODO: Verify Credentials from database here

	// Create JWT Claims
	claims := jwt.MapClaims{
		"username": creds.Username,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key
	authHeader := r.Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	secretKey := parts[1]
	// TODO: verify better

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		message := fmt.Sprintf("Failed to generate JWT: %v", err)
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}

	statusCode := http.StatusOK
	utils.SendMessageResponse(w, tokenString, statusCode)
}

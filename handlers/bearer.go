package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/swarnimcodes/bitespeed-backend-task/utils"
)

// generate a cryptographically safe token
func GenerateBearerToken(w http.ResponseWriter, r *http.Request) {

	const tokenLength = 64

	tokenBytes := make([]byte, tokenLength)

	// read random bytes from the byte slice
	_, err := rand.Read(tokenBytes)
	if err != nil {
		message := fmt.Sprintf("Could not get random bytes: %v", err)
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	statusCode := http.StatusOK
	utils.SendMessageResponse(w, token, statusCode)
}

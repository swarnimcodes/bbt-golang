package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/swarnimcodes/bitespeed-backend-task/utils"
)

func Auth(next http.Handler) http.Handler {
	bearerToken := os.Getenv("BEARER_TOKEN")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			message := "No authorization header sent"
			statusCode := http.StatusUnauthorized
			utils.SendErrorResponse(w, message, statusCode)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			message := "Invalid Authorization header format"
			statusCode := http.StatusUnauthorized
			utils.SendErrorResponse(w, message, statusCode)
			return
		}
		authToken := parts[1]
		if authToken != bearerToken {
			message := "Invalid bearer token sent"
			statusCode := http.StatusUnauthorized
			utils.SendErrorResponse(w, message, statusCode)
			return
		}
		next.ServeHTTP(w, r)
	})
}

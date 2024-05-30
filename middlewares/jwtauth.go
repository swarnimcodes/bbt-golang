package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/swarnimcodes/bitespeed-backend-task/utils"
)

var secretKey = os.Getenv("BEARER_TOKEN")

func JwtAuth(next http.Handler) http.Handler {
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

		tokenString := parts[1]

		// // TODO: fix secret key
		secretKey := os.Getenv("BEARER_TOKEN")
		if secretKey == "" {
			message := "Secret key not set"
			statusCode := http.StatusInternalServerError
			utils.SendErrorResponse(w, message, statusCode)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil || !token.Valid {
			fmt.Println(token)
			message := "Invalid JWT sent"
			statusCode := http.StatusUnauthorized
			utils.SendErrorResponse(w, message, statusCode)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			message := "Invalid JWT claims"
			statusCode := http.StatusUnauthorized
			utils.SendErrorResponse(w, message, statusCode)
			return
		}

		expTime, ok := claims["exp"].(float64)
		if !ok {
			message := "Invalid JWT claims. Missing `exp`."
			statusCode := http.StatusUnauthorized
			utils.SendErrorResponse(w, message, statusCode)
			return
		}
		expDateTime := time.Unix(int64(expTime), 0)
		remainingDuration := time.Until(expDateTime)
		log.Printf("Token will expire at: %s\n", expDateTime)
		log.Printf("Remaining time until token expiration (seconds): %f\n", remainingDuration.Seconds())
		next.ServeHTTP(w, r)
	})
}

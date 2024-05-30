package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/swarnimcodes/bitespeed-backend-task/utils"
)

func PrintHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headersJSON, err := json.MarshalIndent(r.Header, "", "  ")
		if err != nil {
			message := fmt.Sprintf("Error marshalling headers to JSON: %v\n", err)
			statusCode := http.StatusInternalServerError
			utils.SendErrorResponse(w, message, statusCode)
			return
		}
		fmt.Printf("Headers: %s\n", headersJSON)
		next.ServeHTTP(w, r)
	})
}

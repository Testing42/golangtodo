package handlers

import (
	"net/http"
	"os"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")

		// 2. Get the key from the environment variable (defined in .env or System)
		expectedKey := os.Getenv("API_KEY")

		// 3. Compare the header to the environment variable
		// We also check if expectedKey is empty to prevent access if you forgot to set it
		if expectedKey == "" || apiKey != expectedKey {
			// UPDATED: Using the professional JSON error helper
			sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

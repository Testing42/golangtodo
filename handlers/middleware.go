package handlers

import (
	"encoding/json"
	"net/http"
	"os" // 1. Added this to read from the OS environment
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")

		// 2. Get the key from the environment variable we set in PowerShell
		expectedKey := os.Getenv("API_KEY")

		// 3. Compare the header to the environment variable
		// We also check if expectedKey is empty to prevent access if you forgot to set it
		if expectedKey == "" || apiKey != expectedKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		next(w, r)
	}
}

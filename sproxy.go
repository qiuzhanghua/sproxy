package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func proxyHandler(target *url.URL, apiKey string) http.Handler {
	// Create a reverse proxy to the target
	reverseProxy := httputil.NewSingleHostReverseProxy(target)

	// Modify the director to include the API token if needed
	originalDirector := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Optionally, add the API token to the request headers
		// req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the API token from headers
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			http.Error(w, "Unauthorized: Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, prefix)
		if token != apiKey {
			http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
			return
		}

		// Optionally, you can remove the API token header before forwarding
		r.Header.Del("Authorization")

		// Serve the request using the reverse proxy
		reverseProxy.ServeHTTP(w, r)
	})
}

func main() {
	// Target server URL
	target := "http://localhost"

	// Parse the target URL
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}

	// Define the API token
	apiKey := "my-ollama-api-key"

	// Create the proxy handler
	handler := proxyHandler(targetURL, apiKey)

	// Start the HTTP server
	fmt.Println("Starting proxy server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

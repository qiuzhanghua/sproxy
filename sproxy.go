package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func proxyHandler(target *url.URL, token string) http.Handler {
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
		clientToken := r.Header.Get("X-API-Token")
		if clientToken != token {
			http.Error(w, "Unauthorized: Invalid or missing API token", http.StatusUnauthorized)
			return
		}

		// Optionally, you can remove the API token header before forwarding
		r.Header.Del("X-API-Token")

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
	apiToken := "your-secure-api-token"

	// Create the proxy handler
	handler := proxyHandler(targetURL, apiToken)

	// Start the HTTP server
	fmt.Println("Starting proxy server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

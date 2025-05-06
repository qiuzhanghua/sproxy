package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var RedisClient *redis.Client

func init() {
	// Initialize Redis client
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Replace with your Redis server address
		Password: "",               // Replace with your Redis password if any
		DB:       0,                // Use default DB
	})

	// Ping the Redis server to ensure it's reachable
	pong, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis:", pong)
}

func validateAPIKey(ctx context.Context, client *redis.Client, token string) (string, error) {
	// In this example, we'll assume the token is the key and the value is the user ID
	// Adjust the key and value structure as per your Redis schema
	userID, err := client.Get(ctx, token).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("API key not found")
	} else if err != nil {
		return "", fmt.Errorf("redis error: %v", err)
	}

	// Optionally, check for expiration or other metadata
	// For example, if you have a TTL set on the key, you can check if it's still valid
	ttl := client.TTL(ctx, token).Val()
	if ttl <= 0 {
		return "", fmt.Errorf("API key has expired")
	}

	return userID, nil
}

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

		// Validate the token against Redis
		_, err := validateAPIKey(ctx, RedisClient, token)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid or expired API key", http.StatusUnauthorized)
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

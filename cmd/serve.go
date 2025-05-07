package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var withStatic = false
var staticMap = make(map[string]string)

var withRedis = false

var ctx = context.Background()
var RedisClient *redis.Client

func init() {
	s, ok := os.LookupEnv("SECURE_PROXY_WITH_STATIC")
	withStatic = ok && !(s == "0" || strings.ToLower(s) == "false")
	if withStatic {
		initStatic()
	}
	s, ok = os.LookupEnv("SECURE_PROXY_WITH_REDIS")
	withRedis = ok && !(s == "0" || strings.ToLower(s) == "false")
	if withRedis {
		initRedis()
	} else {
		fmt.Println("Redis is not enabled. Running without Redis.")
	}
}

func initStatic() {
	s, ok := os.LookupEnv("SECURE_PROXY_STATIC_MAP")
	if !ok {
		fmt.Println("SECURE_PROXY_STATIC_MAP not set. Running without static.")
		return
	}
	// Parse the static map from the environment variable
	// Example: SECURE_PROXY_STATIC_MAP=key1=user1,key2=user2
	entries := strings.Split(s, ",")
	for _, entry := range entries {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Invalid entry in SECURE_PROXY_STATIC_MAP: %s\n", entry)
			continue
		}
		path := strings.TrimSpace(parts[0])
		target := strings.TrimSpace(parts[1])
		staticMap[path] = target
	}
	fmt.Println("Static map initialized:", staticMap)
}

func validateStatic(token string) (string, error) {
	if user, ok := staticMap[token]; ok {
		return user, nil
	}
	return "", fmt.Errorf("API key not found from static token")
}

func initRedis() {
	redisUrl, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		fmt.Println("REDIS_URL not set. Running without Redis.")
		return
	}
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		fmt.Printf("Failed to parse Redis URL: %v\n", err)
		return
	}

	// Initialize Redis client
	RedisClient = redis.NewClient(opt)

	// Ping the Redis server to ensure it's reachable
	pong, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis:", pong)
}

func validateAPIKey(ctx context.Context, client *redis.Client, token string) (string, error) {
	// In this example, we'll assume the token is the key and the value is the user
	// Adjust the key and value structure as per your Redis schema
	user, err := client.Get(ctx, token).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("API key not found from Redis")
	} else if err != nil {
		return "", fmt.Errorf("redis error: %v", err)
	}

	// Optionally, check for expiration or other metadata
	// For example, if you have a TTL set on the key, you can check if it's still valid
	ttl := client.TTL(ctx, token).Val()
	if ttl <= 0 {
		return "", fmt.Errorf("API key has expired")
	}

	return user, nil
}

func proxyHandler(target *url.URL) http.Handler {
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

		var user string = ""

		// Validate the token against static map
		if user == "" && withStatic {
			if target, err := validateStatic(token); err == nil {
				user = target
			}
		}

		if user == "" && withRedis {
			if target, err := validateAPIKey(ctx, RedisClient, token); err == nil {
				user = target
			}
		}

		if user == "" {
			http.Error(w, "Unauthorized: Invalid or expired API key", http.StatusUnauthorized)
			return
		}

		// Optionally, you can remove the API token header before forwarding
		r.Header.Del("Authorization")

		// Serve the request using the reverse proxy
		reverseProxy.ServeHTTP(w, r)
	})
}

// ServeCmd represents the serve command
var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "start secure proxy server",
	Long:  `start secure proxy server`,
	Run: func(cmd *cobra.Command, args []string) {
		// Target server URL
		target, ok := os.LookupEnv("SECURE_PROXY_TARGET")
		if !ok {
			log.Fatal("SECURE_PROXY_TARGET environment variable is required")
		}

		port, ok := os.LookupEnv("SECURE_PROXY_PORT")
		if !ok {
			log.Fatal("SECURE_PROXY_PORT environment variable is required")
		}

		// Parse the target URL
		targetURL, err := url.Parse(target)
		if err != nil {
			log.Fatalf("Invalid target URL: %v", err)
		}

		// Create the proxy handler
		handler := proxyHandler(targetURL)

		addr := fmt.Sprintf(":%s", port)
		// Start the HTTP server
		fmt.Println("Starting proxy server on ", addr)
		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

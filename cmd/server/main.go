package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/celiojsf/pos-challenge-rate-limiter/internal/config"
	"github.com/celiojsf/pos-challenge-rate-limiter/internal/limiter"
	"github.com/celiojsf/pos-challenge-rate-limiter/internal/middleware"
	"github.com/celiojsf/pos-challenge-rate-limiter/internal/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage (Redis)
	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	store, err := storage.NewRedisStorage(redisAddr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to initialize Redis storage: %v", err)
	}
	defer store.Close()

	log.Println("Connected to Redis successfully")

	// Initialize rate limiter
	rateLimiter := limiter.NewRateLimiter(limiter.Config{
		Storage:       store,
		IPLimit:       cfg.RateLimitIP,
		TokenLimit:    cfg.RateLimitToken,
		BlockDuration: time.Duration(cfg.BlockDuration) * time.Second,
		TokenLimits:   cfg.TokenLimits,
	})

	// Initialize middleware
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(rateLimiter)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(rateLimiterMiddleware.Handle)

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Rate Limiter API", "status": "ok"}`))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})

	r.Get("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Test endpoint", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Create server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port 8080...")
		log.Printf("Rate Limiter Configuration:")
		log.Printf("  - IP Limit: %d requests/second", cfg.RateLimitIP)
		log.Printf("  - Token Limit: %d requests/second", cfg.RateLimitToken)
		log.Printf("  - Block Duration: %d seconds", cfg.BlockDuration)
		log.Printf("  - Custom Token Limits: %v", cfg.TokenLimits)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

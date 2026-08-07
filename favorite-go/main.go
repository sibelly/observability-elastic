package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	redis *redis.Client
}

type Favorite struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func main() {
	redisHost := getenv("REDIS_HOST", "redis:6379")

	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("redis unavailable at startup: %v", err)
	}

	app := &App{
		redis: rdb,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", app.home)
	mux.HandleFunc("GET /health", app.health)
	mux.HandleFunc("GET /favorites", app.listFavorites)
	mux.HandleFunc("POST /favorites", app.createFavorite)

	// endpoints propositalmente ruins para observabilidade
	mux.HandleFunc("GET /slow", app.slow)
	mux.HandleFunc("GET /error", app.fail)

	server := &http.Server{
		Addr:              ":5000",
		Handler:           requestLogger(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("favorite-go listening on :5000")

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func (app *App) home(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "favorite-go",
		"message": "AIBIS observability lab",
	})
}

func (app *App) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := app.redis.Ping(ctx).Err(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"redis":  err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func (app *App) listFavorites(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	values, err := app.redis.HGetAll(ctx, "favorites").Result()
	if err != nil {
		http.Error(w, "failed to read favorites", http.StatusInternalServerError)
		return
	}

	favorites := make([]Favorite, 0, len(values))

	for id, title := range values {
		favorites = append(favorites, Favorite{
			ID:    id,
			Title: title,
		})
	}

	writeJSON(w, http.StatusOK, favorites)
}

func (app *App) createFavorite(w http.ResponseWriter, r *http.Request) {
	var favorite Favorite

	if err := json.NewDecoder(r.Body).Decode(&favorite); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if favorite.ID == "" || favorite.Title == "" {
		http.Error(w, "id and title are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := app.redis.HSet(
		ctx,
		"favorites",
		favorite.ID,
		favorite.Title,
	).Err(); err != nil {
		http.Error(w, "failed to save favorite", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, favorite)
}

func (app *App) slow(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1500 * time.Millisecond)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "that was intentionally slow",
	})
}

func (app *App) fail(w http.ResponseWriter, r *http.Request) {
	http.Error(
		w,
		"simulated internal server error",
		http.StatusInternalServerError,
	)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf(
			"%s %s duration=%s",
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func getenv(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}

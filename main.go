package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/smail1111/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	platform       string
	queries        *database.Queries
}

type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

func main() {
	godotenv.Load()

	DB_URL := os.Getenv("DB_URL")
	PLATFORM := os.Getenv("PLATFORM")

	db, er := sql.Open("postgres", DB_URL)

	if er != nil {
		fmt.Println(er.Error())
	} else {

		config := apiConfig{
			queries:  database.New(db),
			platform: PLATFORM,
		}

		serveMux := http.NewServeMux()

		serveMux.HandleFunc("GET /app", config.middlewareIncrement(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
		serveMux.HandleFunc("GET /api/healthz", handleHealthz)
		serveMux.HandleFunc("GET /admin/metrics", config.handleMetrics)
		serveMux.HandleFunc("POST /admin/reset", config.handleReset)
		serveMux.HandleFunc("POST /api/users", config.handleCreateUser)
		serveMux.HandleFunc("POST /api/chirps", config.handleCreateChirp)
		serveMux.HandleFunc("GET /api/chirps", config.handleGetChirps)
		serveMux.HandleFunc("GET /api/chirps/{id}", config.handleGetChirp)
		serveMux.HandleFunc("POST /api/login", config.handleLogin)

		server := &http.Server{
			Addr:    ":8080",
			Handler: serveMux,
		}

		if er := server.ListenAndServe(); er != nil {
			fmt.Println(er.Error())
		}
	}
}

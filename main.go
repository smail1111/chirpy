package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	config := apiConfig{}

	serveMux := http.NewServeMux()

	serveMux.HandleFunc("GET /app", config.middlewareIncrement(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", handleHealthz)
	serveMux.HandleFunc("GET /api/metrics", config.handleMetrics)
	serveMux.HandleFunc("POST /api/reset", config.handleReset)

	server := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	if er := server.ListenAndServe(); er != nil {
		fmt.Println(er.Error())
	}
}

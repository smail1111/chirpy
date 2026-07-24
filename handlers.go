package main

import (
	"fmt"
	"net/http"
)

func handleHealthz(rs http.ResponseWriter, rq *http.Request) {
	rs.Header().Set("Cintent-Type", "text/plain; charset=utf-8")
	rs.WriteHeader(200)
	rs.Write([]byte("OK"))
}

func (cfg *apiConfig) handleMetrics(rs http.ResponseWriter, rq *http.Request) {
	rs.Header().Set("Cintent-Type", "text/plain; charset=utf-8")
	rs.WriteHeader(200)
	rs.Write(fmt.Appendf([]byte{}, "Hits: %d", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handleReset(rs http.ResponseWriter, rq *http.Request) {
	rs.Header().Set("Cintent-Type", "text/plain; charset=utf-8")
	rs.WriteHeader(200)
	cfg.fileserverHits.Swap(0)
	rs.Write([]byte("RESET"))
}

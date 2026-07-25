package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func handleHealthz(rs http.ResponseWriter, rq *http.Request) {
	rs.Header().Set("Cintent-Type", "text/plain; charset=utf-8")
	rs.WriteHeader(200)
	rs.Write([]byte("OK"))
}

func (cfg *apiConfig) handleMetrics(rs http.ResponseWriter, rq *http.Request) {
	rs.Header().Set("Content-Type", "text/html")
	rs.WriteHeader(200)
	rs.Write(fmt.Appendf([]byte{}, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handleReset(rs http.ResponseWriter, rq *http.Request) {
	if cfg.platform != "dev" {
		returnErrorResponse(rs, 403, "Forbidden")
	} else {
		cfg.fileserverHits.Swap(0)
		cfg.queries.Reset(rq.Context())
		returnJsonResponse(rs, 200, struct {
			Reset bool `json:"reset"`
		}{true})
	}
}

func handleValidateChirp(rs http.ResponseWriter, rq *http.Request) {
	type Chirp struct {
		Body string `json:"body"`
	}

	chirp := Chirp{}
	decoder := json.NewDecoder(rq.Body)

	er := decoder.Decode(&chirp)
	if er != nil {
		returnErrorResponse(rs, 400, er.Error())
	} else {
		if strings.Count((chirp.Body), "")-1 > 140 {
			returnErrorResponse(rs, 400, "Chirp is too long")
		} else {
			returnJsonResponse(rs, 200, struct {
				CleanedBody string `json:"cleaned_body"`
			}{censorBadWords(chirp.Body)})
		}
	}
}

func (cfg *apiConfig) handleCreateUser(rs http.ResponseWriter, rq *http.Request) {
	type Input struct {
		Email string `json:"email"`
	}

	input := Input{}
	decoder := json.NewDecoder(rq.Body)

	er := decoder.Decode(&input)
	if er != nil {
		returnErrorResponse(rs, 400, er.Error())
	} else {
		user, er := cfg.queries.CreateUser(rq.Context(), input.Email)
		if er != nil {
			returnErrorResponse(rs, 400, er.Error())
		} else {
			returnJsonResponse(rs, 201, User{
				user.ID,
				user.CreatedAt,
				user.UpdatedAt,
				user.Email,
			})
		}
	}
}

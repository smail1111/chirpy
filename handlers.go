package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/smail1111/chirpy/internal/auth"
	"github.com/smail1111/chirpy/internal/database"
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

func (cfg *apiConfig) handleCreateUser(rs http.ResponseWriter, rq *http.Request) {
	type Input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	input := Input{}
	decoder := json.NewDecoder(rq.Body)

	er := decoder.Decode(&input)
	if er != nil {
		returnErrorResponse(rs, 400, er.Error())
	} else {
		hashed_password, er := auth.HashPassword(input.Password)
		if er != nil {
			returnErrorResponse(rs, 400, er.Error())
		} else {
			user, er := cfg.queries.CreateUser(rq.Context(), database.CreateUserParams{
				Email:          input.Email,
				HashedPassword: hashed_password,
			})
			if er != nil {
				returnErrorResponse(rs, 400, er.Error())
			} else {
				returnJsonResponse(rs, 201, convertUser(user))
			}
		}
	}
}

func (cfg *apiConfig) handleCreateChirp(rs http.ResponseWriter, rq *http.Request) {
	type Input struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	input := Input{}
	decoder := json.NewDecoder(rq.Body)
	er := decoder.Decode(&input)
	if er != nil {
		returnErrorResponse(rs, 400, er.Error())
	} else {
		if strings.Count((input.Body), "")-1 > 140 {
			returnErrorResponse(rs, 400, "Chirp is too long")
		} else {
			chirp, er := cfg.queries.CreateChirp(rq.Context(), database.CreateChirpParams{
				Body:   censorBadWords(input.Body),
				UserID: input.UserID,
			})
			if er != nil {
				returnErrorResponse(rs, 400, er.Error())
			} else {
				returnJsonResponse(rs, 201, convertChirp(chirp))
			}
		}
	}
}

func (cfg *apiConfig) handleGetChirps(rs http.ResponseWriter, rq *http.Request) {
	chirps, er := cfg.queries.GetChirps(rq.Context())
	if er != nil {
		returnErrorResponse(rs, 400, er.Error())
	} else {
		json_chirps := []Chirp{}
		for _, chirp := range chirps {
			json_chirps = append(json_chirps, convertChirp(chirp))
		}
		returnJsonResponse(rs, 200, json_chirps)
	}
}

func (cfg *apiConfig) handleGetChirp(rs http.ResponseWriter, rq *http.Request) {
	id := rq.PathValue("id")
	chirp, er := cfg.queries.GetChirp(rq.Context(), id)
	if er != nil {
		returnErrorResponse(rs, 404, er.Error())
	} else {
		returnJsonResponse(rs, 200, convertChirp(chirp))
	}
}

func (cfg *apiConfig) handleLogin(rs http.ResponseWriter, rq *http.Request) {
	type Input struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	input := Input{}
	decoder := json.NewDecoder(rq.Body)

	er := decoder.Decode(&input)
	if er != nil {
		returnErrorResponse(rs, 400, er.Error())
	} else {
		user, er := cfg.queries.GetUserByEmail(rq.Context(), input.Email)
		if er != nil {
			returnErrorResponse(rs, 400, er.Error())
		} else {
			matched, er := auth.CheckPasswordHash(input.Password, user.HashedPassword)
			if er != nil {
				returnErrorResponse(rs, 400, er.Error())
			} else {
				if matched {
					returnJsonResponse(rs, 200, convertUser(user))
				} else {
					returnErrorResponse(rs, 401, "Password Incorrect")
				}
			}
		}
	}
}

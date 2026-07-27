package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
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
	rs.Write(fmt.Appendf([]byte{},
		"<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>",
		cfg.fileserverHits.Load(),
	))
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
			returnErrorResponse(rs, 500, er.Error())
		} else {
			user, er := cfg.queries.CreateUser(rq.Context(), database.CreateUserParams{
				Email:          input.Email,
				HashedPassword: hashed_password,
			})
			if er != nil {
				returnErrorResponse(rs, 500, er.Error())
			} else {
				returnJsonResponse(rs, 201, convertUser(user))
			}
		}
	}
}

func (cfg *apiConfig) handleCreateChirp(rs http.ResponseWriter, rq *http.Request, user database.User) {
	type Input struct {
		Body string `json:"body"`
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
				UserID: user.ID,
			})
			if er != nil {
				returnErrorResponse(rs, 500, er.Error())
			} else {
				returnJsonResponse(rs, 201, convertChirp(chirp))
			}
		}
	}
}

func (cfg *apiConfig) handleGetChirps(rs http.ResponseWriter, rq *http.Request) {
	var chirps []database.Chirp
	var er error

	if author_id := rq.URL.Query().Get("author_id"); author_id != "" {
		chirps, er = cfg.queries.GetChirpsByAuthorID(rq.Context(), author_id)
	} else {
		chirps, er = cfg.queries.GetChirps(rq.Context())
	}

	slices.SortFunc(chirps, func(a, b database.Chirp) int {
		return strings.Compare(a.CreatedAt.String(), b.CreatedAt.String())
	})

	if rq.URL.Query().Get("sort") == "desc" {
		slices.Reverse(chirps)
	}

	if er != nil {
		returnErrorResponse(rs, 500, er.Error())
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
			returnErrorResponse(rs, 404, er.Error())
		} else {
			matched, er := auth.CheckPasswordHash(input.Password, user.HashedPassword)
			if er != nil {
				returnErrorResponse(rs, 500, er.Error())
			} else if matched {
				token, er := auth.MakeJWT(uuid.MustParse(user.ID), cfg.secret, time.Second*3600)
				if er != nil {
					returnErrorResponse(rs, 500, er.Error())
				} else {
					refreshToken := auth.MakeRefreshToken()

					_, er := cfg.queries.CreateRefreshToken(rq.Context(), database.CreateRefreshTokenParams{
						Token:  refreshToken,
						UserID: user.ID,
					})

					if er != nil {
						returnErrorResponse(rs, 500, er.Error())
					} else {
						returnJsonResponse(rs, 200, struct {
							User
							Token        string `json:"token"`
							RefreshToken string `json:"refresh_token"`
						}{
							User:         convertUser(user),
							Token:        token,
							RefreshToken: refreshToken,
						})
					}
				}
			} else {
				returnErrorResponse(rs, 401, "Password Incorrect")
			}
		}
	}
}

func (cfg *apiConfig) handleRefresh(rs http.ResponseWriter, rq *http.Request) {
	headerToken, er := auth.GetBearerToken(rq.Header)
	if er != nil {
		returnErrorResponse(rs, 401, er.Error())
		return
	}

	token, er := cfg.queries.GetRefreshToken(rq.Context(), headerToken)
	if er != nil || token.RevokedAt.Valid == true {
		returnErrorResponse(rs, 401, "Unauthorized")
		return
	}

	user, er := cfg.queries.GetUserFromRefreshToken(rq.Context(), token.Token)
	if er != nil {
		returnErrorResponse(rs, 404, er.Error())
		return
	}

	accessToken, er := auth.MakeJWT(uuid.MustParse(user.ID), cfg.secret, time.Second*3600)
	if er != nil {
		returnErrorResponse(rs, 500, er.Error())
		return
	}

	returnJsonResponse(rs, 200, struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handleRevoke(rs http.ResponseWriter, rq *http.Request) {
	headerToken, er := auth.GetBearerToken(rq.Header)
	if er != nil {
		returnErrorResponse(rs, 401, er.Error())
		return
	}

	er = cfg.queries.RevokeRefreshToken(rq.Context(), headerToken)
	if er != nil {
		returnErrorResponse(rs, 404, er.Error())
		return
	}

	rs.WriteHeader(204)
}

func (cfg *apiConfig) handleUpdateUser(rs http.ResponseWriter, rq *http.Request, user database.User) {
	type Input struct {
		Email    string `json:"email"`
		Password string `josn:"password"`
	}

	input := Input{}
	decoder := json.NewDecoder(rq.Body)

	er := decoder.Decode(&input)
	if er != nil {
		returnErrorResponse(rs, 400, er.Error())
		return
	}

	hashedPassword, er := auth.HashPassword(input.Password)
	if er != nil {
		returnErrorResponse(rs, 500, er.Error())
		return
	}

	updatedUser, er := cfg.queries.UpdateUser(rq.Context(), database.UpdateUserParams{
		Email:          input.Email,
		HashedPassword: hashedPassword,
		ID:             user.ID,
	})

	if er != nil {
		returnErrorResponse(rs, 500, er.Error())
		return
	}

	returnJsonResponse(rs, 200, convertUser(updatedUser))
}

func (cfg *apiConfig) handleDeleteChirp(rs http.ResponseWriter, rq *http.Request, user database.User) {
	id := rq.PathValue("id")

	chirp, er := cfg.queries.GetChirp(rq.Context(), id)
	if er != nil {
		returnErrorResponse(rs, 404, er.Error())
		return
	}

	if chirp.UserID != user.ID {
		returnErrorResponse(rs, 403, "Forbidden")
		return
	}

	er = cfg.queries.DeleteChirp(rq.Context(), chirp.ID)
	if er != nil {
		returnErrorResponse(rs, 500, er.Error())
		return
	}

	rs.WriteHeader(204)
}

func (cfg *apiConfig) handleUserUpgraded(rs http.ResponseWriter, rq *http.Request) {
	apiKey, er := auth.GetAPIKey(rq.Header)
	if er != nil || apiKey != cfg.polka_key {
		returnErrorResponse(rs, 401, "Unauthorized")
		return
	}

	type Input struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	input := Input{}
	decoder := json.NewDecoder(rq.Body)

	er = decoder.Decode(&input)
	if er != nil {
		returnErrorResponse(rs, 400, er.Error())
		return
	}

	if input.Event == "user.upgraded" {
		_, er = cfg.queries.UpgradeUser(rq.Context(), input.Data.UserID)
		if er != nil {
			returnErrorResponse(rs, 404, er.Error())
			return
		}
	}

	rs.WriteHeader(204)
}

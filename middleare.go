package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/smail1111/chirpy/internal/auth"
	"github.com/smail1111/chirpy/internal/database"
)

type errorResponse struct {
	Error string `json:"error"`
}

func (cfg *apiConfig) middlewareIncrement(handler http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(rs http.ResponseWriter, rq *http.Request) {
		cfg.fileserverHits.Add(1)

		handler.ServeHTTP(rs, rq)
	}
}

func returnErrorResponse(rs http.ResponseWriter, code int, er string) {
	rs.Header().Set("Content-Type", "application/json")

	data, _ := json.Marshal(errorResponse{
		Error: er,
	})

	rs.WriteHeader(code)
	rs.Write(data)
}

func returnJsonResponse(rs http.ResponseWriter, code int, payload any) {
	rs.Header().Set("Content-Type", "application/json")

	data, er := json.Marshal(payload)
	if er != nil {
		returnErrorResponse(rs, 400, "Malformed Input")
	} else {
		rs.WriteHeader(code)
		rs.Write(data)
	}
}

func censorBadWords(text string) string {
	censored := []string{}
	for _, word := range strings.Split(text, " ") {
		if lower := strings.ToLower(word); lower == "kerfuffle" || lower == "sharbert" || lower == "fornax" {
			word = "****"
		}
		censored = append(censored, word)
	}
	return strings.Join(censored, " ")
}

func convertChirp(chirp database.Chirp) Chirp {
	return Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
}

func convertUser(user database.User) User {
	return User{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
		user.IsChirpyRed,
	}
}

func (cfg *apiConfig) middlewareAuthorizeUser(handler func(rs http.ResponseWriter, rq *http.Request,
	user database.User)) func(http.ResponseWriter, *http.Request) {
	return func(rs http.ResponseWriter, rq *http.Request) {
		headerToken, er := auth.GetBearerToken(rq.Header)
		if er != nil {
			returnErrorResponse(rs, 401, er.Error())
			return
		}

		id, er := auth.ValidateJWT(headerToken, cfg.secret)
		if er != nil {
			returnErrorResponse(rs, 401, er.Error())
			return
		}

		user, er := cfg.queries.GetUserByID(rq.Context(), id.String())
		if er != nil {
			returnErrorResponse(rs, 404, er.Error())
			return
		}

		handler(rs, rq, user)
	}
}

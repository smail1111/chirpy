package main

import (
	"encoding/json"
	"net/http"
	"strings"
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

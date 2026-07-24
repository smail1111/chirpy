package main

import "net/http"

func (cfg *apiConfig) middlewareIncrement(handler http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(rs http.ResponseWriter, rq *http.Request) {
		cfg.fileserverHits.Add(1)

		handler.ServeHTTP(rs, rq)
	}
}

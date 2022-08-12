package main

import (
	"encoding/json"
	"fmt"
	"github.com/sang-w0o/gag"
	"log"
	"net/http"
	"time"

	gorillaMux "github.com/gorilla/mux"
)

type sampleResponse struct {
	Message string `json:"message"`
}

func sampleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := json.Marshal(sampleResponse{Message: "sample handler!"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("something went wrong.."))
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}

func samplePathVariableHandler() http.HandlerFunc {

	type sampleResponse struct {
		Message string `json:"message"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := gorillaMux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("something went wrong.."))
		}
		res, err := json.Marshal(sampleResponse{Message: fmt.Sprintf("id: %s", id)})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("something went wrong.."))
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}

func sampleTimingMiddleware() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			end := time.Now()
			log.Printf("request time for %s: %v", r.URL.Path, end.Sub(start))
		})
	}
}

func TerribleSecurityProvider(password string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Secret-Password") != password {
				res, err := json.Marshal(sampleResponse{Message: "authentication failed"})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("something went wrong.."))
					return
				} else {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write(res)
					return
				}
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}

func main() {
	cfg := gag.Config{Port: 8080}
	g := gag.NewGag(cfg)
	g.Conditions().
		Path("/a").Method(http.MethodGet).Route(&gag.RouteRequest{Url: "/route-to", HttpMethod: http.MethodGet}, g).
		Path("/b").Method(http.MethodGet).HasHeader("X-Header-Key").Route(&gag.RouteRequest{Url: "/route-to", HttpMethod: http.MethodGet}, g).
		Path("/c").Method(http.MethodGet).HasHeaderValue("X-Key", "someValue").HandlerFunc(sampleHandler(), g).
		Path("/d").Method(http.MethodPost).Route(&gag.RouteRequest{Url: "/route-to-handle", HttpMethod: http.MethodPost}, g).
		Path("/e").Middlewares(TerribleSecurityProvider("some"), sampleTimingMiddleware()).HandlerFunc(sampleHandler(), g).
		Path("/f").Middlewares(sampleTimingMiddleware(), TerribleSecurityProvider("some")).HandlerFunc(sampleHandler(), g).
		Path("/demo").Middlewares(sampleTimingMiddleware()).Method(http.MethodPost).Route(&gag.RouteRequest{
		Url:             "http://127.0.0.1:8081/demo",
		HttpMethod:      http.MethodPost,
		Timeout:         2 * time.Second,
		PassRequestBody: true,
	}, g).
		Path("/user").HandlerFunc(sampleHandler(), g).
		Path("/this-is-path/{id}").HandlerFunc(samplePathVariableHandler(), g)
	err := g.Serve()
	if err != nil {
		panic(err)
	}
}

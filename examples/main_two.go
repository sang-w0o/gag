package main

import (
	"github.com/sang-w0o/gag"
	"net/http"
)

func main() {
	cfg := gag.Config{Port: 8080}
	g := gag.NewGag(cfg)
	g.Conditions().
		Path("/a").Method(http.MethodGet).HandlerFunc(sampleHandler(), g).
		Path("/b").Method(http.MethodPost).HandlerFunc(sampleHandler(), g).
		Path("/c").Method(http.MethodGet).HasHeader("X-Header-Key").HandlerFunc(sampleHandler(), g).
		Path("/d").Method(http.MethodGet).HasHeaderValue("X-Key", "someValue").HandlerFunc(sampleHandler(), g).
		Path("/e").Method(http.MethodGet).HasHeader("X-Key-Two").HasHeaderValue("X-Header-Key", "someValue").HandlerFunc(sampleHandler(), g)
	err := g.Serve()
	if err != nil {
		panic(err)
	}
}

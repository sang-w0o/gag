package gag

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"testing"
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

var port uint16
var c *http.Client

func TestMain(m *testing.M) {
	g := NewGag(Config{Port: 8080})
	g.Conditions().
		Path("/a").Method(http.MethodGet).HandlerFunc(sampleHandler(), g).
		Path("/b").Method(http.MethodPost).HandlerFunc(sampleHandler(), g).
		Path("/c").Method(http.MethodGet).HasHeader("X-Key").HandlerFunc(sampleHandler(), g).
		Path("/d").Method(http.MethodGet).HasHeaderValue("X-Key", "someValue").HandlerFunc(sampleHandler(), g).
		Path("/e").Method(http.MethodGet).HasHeader("X-Key").HasHeaderValue("X-Key-Two", "someValue").HandlerFunc(sampleHandler(), g)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := g.Serve()
		defer wg.Done()
		if err != nil {
			panic(err)
		}
	}()

	port = g.port
	c = http.DefaultClient
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestCorrectHttpMethodHandlingSuccess(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/a", port), nil)
	if err != nil {
		t.Errorf("error creating request: %v", err)
		return
	}

	res, err := c.Do(r)
	if err != nil {
		t.Errorf("error doing request: %v", err)
		return
	}

	if err := validateResponse(res, http.StatusOK, `{"message":"sample handler!"}`); err != nil {
		t.Error(err)
		return
	}

	if res.Header["Content-Type"][0] != "application/json" {
		t.Errorf("expected content type %s, got %s", "application/json", res.Header["Content-Type"][0])
		return
	}
}

func TestWrongHttpMethodResponse405(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/b", port), nil)
	if err != nil {
		t.Errorf("error creating request: %v", err)
		return
	}

	res, err := c.Do(r)
	if err != nil {
		t.Errorf("error doing request: %v", err)
		return
	}

	if err := validateResponse(res, http.StatusMethodNotAllowed, "405 method(GET) not allowed"); err != nil {
		t.Error(err)
		return
	}

	if res.Header["Content-Type"][0] != "text/plain; charset=utf-8" {
		t.Errorf("expected content type %s, got %s", "text/plain; charset=utf-8", res.Header["Content-Type"][0])
		return
	}
}

func TestCorrectHeaderHandlingSuccess(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/c", port), nil)
	r.Header.Set("X-Key", "someValue")
	if err != nil {
		t.Errorf("error creating request: %v", err)
		return
	}

	res, err := c.Do(r)
	if err != nil {
		t.Errorf("error doing request: %v", err)
		return
	}

	if err := validateResponse(res, http.StatusOK, `{"message":"sample handler!"}`); err != nil {
		t.Error(err)
		return
	}

	if res.Header["Content-Type"][0] != "application/json" {
		t.Errorf("expected content type %s, got %s", "application/json", res.Header["Content-Type"][0])
		return
	}
}

func TestWrongHeaderResponse400(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/c", port), nil)
	if err != nil {
		t.Errorf("error creating request: %v", err)
		return
	}

	res, err := c.Do(r)
	if err != nil {
		t.Errorf("error doing request: %v", err)
		return
	}

	if err := validateResponse(res, http.StatusBadRequest, "400 header(X-Key) not provided"); err != nil {
		t.Error(err)
		return
	}

	if res.Header["Content-Type"][0] != "text/plain; charset=utf-8" {
		t.Errorf("expected content type %s, got %s", "text/plain; charset=utf-8", res.Header["Content-Type"][0])
		return
	}
}

func TestCorrectHeaderValueHandlingSuccess(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/d", port), nil)
	r.Header.Set("X-Key", "someValue")
	if err != nil {
		t.Errorf("error creating request: %v", err)
		return
	}

	res, err := c.Do(r)
	if err != nil {
		t.Errorf("error doing request: %v", err)
		return
	}

	if err := validateResponse(res, http.StatusOK, `{"message":"sample handler!"}`); err != nil {
		t.Error(err)
		return
	}

	if res.Header["Content-Type"][0] != "application/json" {
		t.Errorf("expected content type %s, got %s", "application/json", res.Header["Content-Type"][0])
		return
	}
}

func TestWrongHeaderValueResponse400(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/d", port), nil)
	r.Header.Set("X-Key", "someWrongValue")
	if err != nil {
		t.Errorf("error creating request: %v", err)
		return
	}

	res, err := c.Do(r)
	if err != nil {
		t.Errorf("error doing request: %v", err)
		return
	}

	if err := validateResponse(res, http.StatusBadRequest, "400 header(X-Key) with value(someValue) not provided"); err != nil {
		t.Error(err)
		return
	}

	if res.Header["Content-Type"][0] != "text/plain; charset=utf-8" {
		t.Errorf("expected content type %s, got %s", "text/plain; charset=utf-8", res.Header["Content-Type"][0])
		return
	}
}

func TestCorrectHeaderAndHeaderValueHandlingSuccess(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/e", port), nil)
	r.Header.Set("X-Key", "someWrongValue")
	r.Header.Set("X-Key-Two", "someValue")
	if err != nil {
		t.Errorf("error creating request: %v", err)
		return
	}

	res, err := c.Do(r)
	if err != nil {
		t.Errorf("error doing request: %v", err)
		return
	}

	if err := validateResponse(res, http.StatusOK, `{"message":"sample handler!"}`); err != nil {
		t.Error(err)
		return
	}

	if res.Header["Content-Type"][0] != "application/json" {
		t.Errorf("expected content type %s, got %s", "application/json", res.Header["Content-Type"][0])
		return
	}
}

func TestWrongHeaderWhenHeaderAndHeaderValueResponse400(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/e", port), nil)
	r.Header.Set("X-Key-Two", "someValue")
	if err != nil {
		t.Errorf("error creating request: %v", err)
		return
	}

	res, err := c.Do(r)
	if err != nil {
		t.Errorf("error doing request: %v", err)
		return
	}

	if err := validateResponse(res, http.StatusBadRequest, "400 header(X-Key) not provided"); err != nil {
		t.Error(err)
		return
	}

	if res.Header["Content-Type"][0] != "text/plain; charset=utf-8" {
		t.Errorf("expected content type %s, got %s", "text/plain; charset=utf-8", res.Header["Content-Type"][0])
		return
	}
}

func TestWrongHeaderValueWhenHeaderAndHeaderValueResponse400(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/e", port), nil)
	r.Header.Set("X-Key", "someWrongValue")
	if err != nil {
		t.Errorf("error creating request: %v", err)
		return
	}

	res, err := c.Do(r)
	if err != nil {
		t.Errorf("error doing request: %v", err)
		return
	}

	if err := validateResponse(res, http.StatusBadRequest, "400 header(X-Key-Two) with value(someValue) not provided"); err != nil {
		t.Error(err)
		return
	}

	if res.Header["Content-Type"][0] != "text/plain; charset=utf-8" {
		t.Errorf("expected content type %s, got %s", "text/plain; charset=utf-8", res.Header["Content-Type"][0])
		return
	}
}

func validateResponse(r *http.Response, statusCode int, body string) error {
	if r.StatusCode != statusCode {
		return fmt.Errorf("expected status code %d, got %d", statusCode, r.StatusCode)
	}

	defer r.Body.Close()

	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	if string(respBody) != body {
		return fmt.Errorf("expected response body %s, got %s", body, string(respBody))
	}

	return nil
}

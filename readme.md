[![Go Reference](https://pkg.go.dev/badge/github.com/sang-w0o/gag.svg)](https://pkg.go.dev/github.com/sang-w0o/gag)

# Gag 

- Gag is an HTTP API Gateway, implemented in Go.  
  Gag provides a declarative interface, and it is highly configurable.

### Supported features

- Handle requests based on path.
- Handle requests based on HTTP method.
- Handle requests that have configured header key.
- Handle requests that have configured header key, along with value.
- Handle requests with your own handler, which implements [API composition pattern](https://microservices.io/patterns/data/api-composition.html)
- Route(redirect) requests to different services.
- Apply middlewares for each request.

### Examples

#### Simple Routing

```go
func main() {
    g := gag.NewGag(gag.Config{Port: 8080})
    g.Conditions.
        Path("/foo").Method(http.MethodGet).Route(&gag.RouteRequest{Url: "http://some.url/route-to", HttpMethod: http.MethodGet}, g)
    err := g.Serve()
    if err != nil {
        panic(err)
    }
}
```

- In the above example, only requests that match `/foo` path, has HTTP GET method will be routed to `http://some.url/route-to`.

#### Simple handling

```go
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

func main() {
    g := gag.NewGag(gag.Config{Port: 8080})
    g.Conditions.
        Path("/bar").Method(http.MethodPost).HandlerFunc(sampleHandler(), g)
    err := g.Serve()
    if err != nil {
        panic(err)
    }
}
```

- In the above example, only requests that match `/bar` path, has HTTP POST method will be handled.

#### Using middlewares

```go
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

func main() {
    g := gag.NewGag(gag.Config{Port: 8080})
	g.Conditions().
		Path("/some-path").Middlewares(sampleTimingMiddleware()).HandlerFunc(sampleHandler(), g)
	err := g.Serve()
	if err != nil {
		panic(err)
	}
}
```

- Since Gag is built on top of [gorilla/mux](https://github.com/gorilla/mux), path supports path variables and much more.   
  Below is an example code using path variable.

```go
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

func main() {
    g := gag.NewGag(gag.Config{Port: 8080})
	g.Conditions().
        Path("/this-is-path/{id}").HandlerFunc(samplePathVariableHandler(), g)
    err := g.Serve()
    if err != nil {
        panic(err)
    }
}
```
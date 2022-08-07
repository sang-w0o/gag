package gag

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	gorillaMux "github.com/gorilla/mux"
)

// Config contains properties for Gag.
type Config struct {
	// Port defines which port number will be used to listen to HTTP requests.
	// When given 0, Gag will start on random available port.
	Port uint16
}

// Gag is a struct that contains all the necessary properties to run Gag.
type Gag struct {
	port       uint16
	l          net.Listener
	s          *http.Server
	conditions []*Condition
	mux        *gorillaMux.Router
	log        logger
}

func (g *Gag) listenHTTP(port uint16) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return errors.New("failed to obtain tcp address")
	}

	g.l = l
	g.port = uint16(tcpAddr.Port)
	g.configureHandler()
	g.log.Println(fmt.Sprintf("gag started on port %d", g.port))
	if err = g.newServer(); err != nil {
		return err
	}
	return nil
}

func (g *Gag) newServer() error {
	if err := g.serve(); err != nil {
		return err
	}
	return nil
}

func (g *Gag) serve() error {
	if err := http.Serve(g.l, g.mux); err != nil {
		fmt.Printf("err in http.Serve(): %s\n", err.Error())
		return err
	}
	return nil
}

// Serve starts an HTTP server.
func (g *Gag) Serve() error {
	if err := g.validateConditions(); err != nil {
		return err
	}
	if err := g.listenHTTP(g.port); err != nil {
		return err
	}
	return nil
}

// NewGag returns a new Gag instance.
func NewGag(cfg Config) *Gag {
	g := Gag{
		port:       cfg.Port,
		conditions: []*Condition{},
		log:        logger{},
	}
	return &g
}

// Conditions is used to add Conditions to Gag.
// For example:
// 	g := NewGag(Config{})
// 	g.Conditions().Path("/foo").Method(http.MethodGet).Route(&gag.RouteRequest{Url: "/route-to", HttpMethod: http.MethodGet}, g)
func (g *Gag) Conditions() *Condition {
	return &Condition{}
}

func (g *Gag) validateConditions() error {
	for _, c := range g.conditions {
		if c.path == "" {
			return errors.New("path cannot be \"\"")
		}
	}
	return nil
}

func (g *Gag) configureHandler() {
	g.log.Println(fmt.Sprintf("total conditions found: %d", len(g.conditions)))
	mux := gorillaMux.NewRouter()
	for _, c := range g.conditions {
		configuredMux := configureMuxHandlers(c)
		mux.Handle(c.path, configuredMux)
		g.log.Println(fmt.Sprintf("path %s registered", c.path))
	}
	g.mux = mux
}

func configureMuxHandlers(c *Condition) *gorillaMux.Router {
	//mux := http.NewServeMux()
	mux := gorillaMux.NewRouter()
	var h http.Handler
	if len(c.middlewares.middlewares) > 0 {
		if c.handlerFunc != nil {
			h = c.middlewares.wrap(c.handlerFunc, h)
		} else {
			h = c.middlewares.wrap(func(w http.ResponseWriter, r *http.Request) {
				client := http.Client{Timeout: c.routeRequest.Timeout}
				if c.routeRequest.PassRequestBody {
					defer r.Body.Close()
					reqBody := r.Body
					req, err := http.NewRequestWithContext(r.Context(), c.routeRequest.HttpMethod, c.routeRequest.Url, reqBody)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					req.Header.Set("Content-Type", "application/json")
					resp, err := client.Do(req)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					defer resp.Body.Close()
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(resp.StatusCode)
					w.Write(bodyBytes)
					return
				} else {
					req, err := http.NewRequestWithContext(r.Context(), c.routeRequest.HttpMethod, c.routeRequest.Url, nil)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					resp, err := client.Do(req)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					defer resp.Body.Close()
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(resp.StatusCode)
					w.Write(bodyBytes)
					return
				}
			}, h)
		}
	} else {
		if c.handlerFunc != nil {
			h = c.handlerFunc
		} else {
			h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				client := http.Client{Timeout: c.routeRequest.Timeout}
				if c.routeRequest.PassRequestBody {
					defer r.Body.Close()
					reqBody := r.Body
					req, err := http.NewRequestWithContext(r.Context(), c.routeRequest.HttpMethod, c.routeRequest.Url, reqBody)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					req.Header.Set("Content-Type", "application/json")
					resp, err := client.Do(req)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					defer resp.Body.Close()
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(resp.StatusCode)
					w.Write(bodyBytes)
					return
				} else {
					req, err := http.NewRequestWithContext(r.Context(), c.routeRequest.HttpMethod, c.routeRequest.Url, nil)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					resp, err := client.Do(req)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					defer resp.Body.Close()
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(resp.StatusCode)
					w.Write(bodyBytes)
					return
				}
			})
		}
	}

	mux.HandleFunc(c.path, func(w http.ResponseWriter, r *http.Request) {
		if (c.httpMethod != "" && c.httpMethod == r.Method) || (c.httpMethod == "") {
			if _, ok := r.Header[c.header]; ok {
				if c.headerValue != nil {
					if values, ok := r.Header[c.headerValue.Key]; hasHeaderValue(c.headerValue.Value, values) && ok {
						h.ServeHTTP(w, r)
					}
				} else {
					h.ServeHTTP(w, r)
				}
			} else if c.header == "" {
				if c.headerValue != nil {
					if values := r.Header[c.headerValue.Key]; hasHeaderValue(c.headerValue.Value, values) {
						h.ServeHTTP(w, r)
					}
				} else {
					h.ServeHTTP(w, r)
				}
			}
		}
	})
	return mux
}

func hasHeaderValue(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

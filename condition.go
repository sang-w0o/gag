package gag

import (
	"net/http"
	"time"
)

// Middleware is a function that will be called before, after, or instead of the handler function.
// Add Middlewares using the Condition.Middlewares() method.
type Middleware func(http.Handler) http.Handler

// middlewareChain contains a list of middlewares to be applied to a handler.
type middlewareChain struct {
	middlewares []Middleware
}

// Condition contains all properties which will be used to determine
// whether a request should be handled by the gag.
type Condition struct {
	// httpMethod represents the HTTP method of the request.
	// If not set, all methods will match.
	// If not empty, requests having the same HTTP method as httpMethod will be handled.
	// Configure httpMethod using Condition.Method() method.
	httpMethod string
	// routeRequest contains all properties about where and how the request will be routed.
	// Only one of routeRequest or handlerFunc can be set per Condition.
	// Configure routeRequest using Condition.Route() method.
	routeRequest *RouteRequest
	// header represents the header key of the request.
	// If not set, header will not be checked.
	// If set, only the requests having the header key same as header will be handled.
	// Configure header using Condition.HasHeader() method.
	header string
	// headerValue represents a key-value pair of request header.
	// If not set, whether the header key exists and has the same value will not be checked.
	// If set, only the request having the same header key along with header value will be handled.
	// headerValue is configured when Condition.HasHeaderValue() method is called.
	headerValue *headerValue
	// path represents the path of the request.
	// It should be always be set, otherwise the request will not be handled.
	// Requests matching the path will be handled.
	path string
	// handlerFunc is the handler function of the request.
	// Only one of handlerFunc or routeRequest can be set per Condition.
	// Configure handlerFunc using Condition.HandlerFunc() method.
	handlerFunc http.HandlerFunc
	// middlewares is a list of middlewares to be applied to the request.
	// Configure middlewares using Condition.Middlewares() method.
	middlewares middlewareChain
}

// RouteRequest contains all properties about where and how the request will be routed.
type RouteRequest struct {
	// Url is the url that the request will be routed to.
	Url string
	// HttpMethod is the HTTP method that will be used to route the request.
	HttpMethod string
	// Timeout is the timeout value of the request, which will be sent to the Url.
	Timeout time.Duration
	// PassRequestBody determines whether the request body will be sent to the Url.
	PassRequestBody bool
}

type headerValue struct {
	Key   string
	Value string
}

// Method sets Condition's httpMethod property.
// If empty or not called, all requests will be handled.
// If not empty, requests having the same HTTP method as httpMethod will be handled.
// Example:
//  g.Condition().Path("/foo").Method(http.MethodGet).Route(...)
func (c *Condition) Method(httpMethod string) *Condition {
	c.httpMethod = httpMethod
	return c
}

// Path sets Condition's path property.
// It should be always be set, otherwise the request will not be handled.
// Requests matching the path will be handled.
// Example:
//  g.Condition().Path("/foo").Route(...)
func (c *Condition) Path(path string) *Condition {
	c.path = path
	return c
}

// HasHeader sets Condition's header property.
// If not set, header will not be checked.
// If set, only the requests having the header key same as header will be handled.
// Example:
//  g.Condition().Path("/foo").HasHeader("X-My-Header").Route(...)
func (c *Condition) HasHeader(header string) *Condition {
	c.header = header
	return c
}

// HasHeaderValue sets Condition's headerValue property.
// If not set, whether the header key exists and has the same value will not be checked.
// If set, only the request having the same header key along with header value will be handled.
// Example:
//  g.Condition().Path("/foo").HasHeaderValue("X-My-Header", "my-value").Route(...)
func (c *Condition) HasHeaderValue(key string, value string) *Condition {
	c.headerValue = &headerValue{key, value}
	return c
}

// Middlewares sets Condition's middlewares property.
// Example:
//  func sampleTimingMiddleware() func(h http.Handler) http.Handler {
//	  return func(h http.Handler) http.Handler {
//		  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			  start := time.Now()
//			  h.ServeHTTP(w, r)
//			  end := time.Now()
//			  log.Printf("request time for %s: %v", r.URL.Path, end.Sub(start))
//		  })
//	  }
//  }
//
//  g.Condition().Path("/foo").Middlewares(sampleTimingMiddleware()).Route(...)
func (c *Condition) Middlewares(middlewares ...Middleware) *Condition {
	c.middlewares = middlewareChain{append(([]Middleware)(nil), middlewares...)}
	return c
}

// Route sets Condition's routeRequest property.
// Only one of routeRequest or handlerFunc can be set per Condition.
// Example:
//  g.Conditions().
//	  Path("/demo").Method(http.MethodPost).Route(&gag.RouteRequest{
//		  Url:             "http://127.0.0.1:8081/demo",
//		  HttpMethod:      http.MethodPost,
//		  Timeout:         2 * time.Second,
//		  PassRequestBody: true,
//	  }, g)
func (c *Condition) Route(routeRequest *RouteRequest, g *Gag) *Condition {
	c.routeRequest = routeRequest
	g.conditions = append(g.conditions, c)
	return &Condition{}
}

// HandlerFunc sets Condition's handlerFunc property.
// Example:
//  func sampleHandler() http.HandlerFunc {
//	  return func(w http.ResponseWriter, r *http.Request) {
//		  w.WriteHeader(http.StatusOK)
//		  w.Write([]byte("result"))
//	  }
//  }
//  g.Condition().Path("/foo").HandlerFunc(sampleHandler())
func (c *Condition) HandlerFunc(handlerFunc http.HandlerFunc, g *Gag) *Condition {
	c.handlerFunc = handlerFunc
	g.conditions = append(g.conditions, c)
	return &Condition{}
}

func (mc middlewareChain) wrap(handlerFunc http.HandlerFunc, h http.Handler) http.Handler {
	if h == nil {
		h = http.NewServeMux()
	}
	for i := range mc.middlewares {
		if i == 0 {
			h = mc.middlewares[i](handlerFunc)
		} else {
			h = mc.middlewares[i](h)
		}
	}
	return h
}

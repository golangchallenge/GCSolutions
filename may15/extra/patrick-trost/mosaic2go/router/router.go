package router

import (
	"net/http"
	"sort"
)

// The Router implements the http.Handler interface and is responsible for mapping
// URL's and request methods to controllers that handle the request.
type Router struct {
	Routes []*Route
}

// ServeHTTP iterates over each registered routes in the order they were registered.
// When a match is found, the request is forwarded to that routes handler.
// StatusNotFound is returned If no route matches.
func (r *Router) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	var keys []int
	var handler http.Handler
	for k := range r.Routes {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		if r.Routes[k].Match(request) {
			handler = r.Routes[k].Handler
			break
		}
	}
	if handler == nil {
		handler = http.NotFoundHandler()
	}
	handler.ServeHTTP(w, request)
}

// Get registers a http.Handler for a URL pattern and request method GET.
func (r *Router) Get(pattern string, handler http.Handler) {
	r.Routes = append(r.Routes, NewRoute(pattern, "GET", handler))
}

// Post registers a http.Handler for a URL pattern and request method POST.
func (r *Router) Post(pattern string, handler http.Handler) {
	r.Routes = append(r.Routes, NewRoute(pattern, "POST", handler))
}

// Put registers a http.Handler for a URL pattern and request method PUT.
func (r *Router) Put(pattern string, handler http.Handler) {
	r.Routes = append(r.Routes, NewRoute(pattern, "PUT", handler))
}

// Delete registers a http.Handler for URL pattern and request method DELETE.
func (r *Router) Delete(pattern string, handler http.Handler) {
	r.Routes = append(r.Routes, NewRoute(pattern, "DELETE", handler))
}

// New creates an new Router pointer.
func New() *Router {
	return &Router{}
}

// Route defines a type that holds all information for a single route.
type Route struct {
	Pattern string
	Method  string
	Handler http.Handler
}

// Match determains whether a route matches a given http request.
func (r *Route) Match(request *http.Request) bool {
	if (request.URL.Path == r.Pattern) && (request.Method == r.Method) {
		return true
	}
	return false
}

// NewRoute creates an new Route pointer.
func NewRoute(pattern string, method string, handler http.Handler) *Route {
	return &Route{
		Pattern: pattern,
		Method:  method,
		Handler: handler,
	}
}

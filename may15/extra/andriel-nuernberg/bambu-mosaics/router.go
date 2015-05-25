package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Router
type Router struct {
	handlers []*Handler
}

// Handler is a http.Handler.
type Handler struct {
	Method string
	Path   string
	Regex  *regexp.Regexp
	http.Handler
}

// NewRouter returns an empty Router.
func NewRouter() *Router {
	return &Router{}
}

// Add register a new Handler to the router with the given method and path.
func (r *Router) Add(method string, path string, handler http.Handler) {
	r.handlers = append(r.handlers, &Handler{method, path, regexfy(path), handler})
}

// Get is a shortcut for register GET handlers.
// The verb HEAD is also registered, which is useful, for example, with the `-I`
// curl option.
func (r *Router) Get(path string, handler http.Handler) {
	r.Add("GET", path, handler)
	r.Add("HEAD", path, handler)
}

// Post is a shortcut for register POST handlers.
func (r *Router) Post(path string, handler http.Handler) {
	r.Add("POST", path, handler)
}

// ServeHTTP makes the routes implement the http.Handler interface.
func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	for _, h := range r.handlers {
		if ok, params := h.matchRequest(req); ok {
			req.URL.RawQuery = params.Encode() + "&" + req.URL.RawQuery
			h.ServeHTTP(res, req)
			return
		}
	}

	http.NotFound(res, req)
}

// matchRequest matches que request request with the registered by the route.
func (h *Handler) matchRequest(req *http.Request) (bool, url.Values) {
	// Not the same HTTP verb.
	if h.Method != req.Method || !h.Regex.MatchString(req.URL.Path) {
		return false, nil
	}

	// Static route. No params.
	if h.Path == req.URL.Path {
		return true, nil
	}

	// Compare the route with the request path.
	matches := h.Regex.FindStringSubmatch(strings.TrimRight(req.URL.Path, "/"))
	if len(matches) <= 1 {
		return false, nil
	} else {
		matches = matches[1:]
	}

	values := make(url.Values)
	params := strings.Split(h.Path, "/")

	j := 0
	for _, p := range params {
		if strings.HasPrefix(p, ":") {
			values.Add(p[1:], matches[j])
			j++
		}
	}

	return true, values
}

// regexfy compiles the registered handler path to a regex object. It permits
// collect the parameters in the URL.
//
// For a path like "/blog/:slug", the compiled regex will be "/blog/(.*)",
// wich will make the "slug" parameter available inside the handler.
func regexfy(path string) *regexp.Regexp {
	parts := strings.Split(path, "/")

	for i, p := range parts {
		if strings.HasPrefix(p, ":") {
			parts[i] = "(.*)"
		}
	}

	pattern := strings.Join(parts, "/")
	regex, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	return regex
}

// Listen register the router to the http.Handler root and make the HTTP server
// available in the given port.
func (r *Router) Listen(port int) {
	http.Handle("/", r)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

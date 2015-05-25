package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ptrost/mosaic2go/test"
)

func TestRouterServeHTTP(t *testing.T) {
	router := New()
	handler1 := NewTestHandler()
	handler2 := NewTestHandler()
	router.Get("/", handler1)
	router.Get("/", handler2)
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	test.Assert("Router.ServeHTTP", http.StatusOK, response.Code, t)
	if handler1.IsCalled == false {
		t.Errorf("First registered handler was not matched first.")
	}
	if handler2.IsCalled == true {
		t.Errorf("Last registered handler was matched but shouldn't.")
	}
}

func TestRouterGet(t *testing.T) {
	router := New()
	handler := NewTestHandler()
	router.Get("/test", handler)
	request, _ := http.NewRequest("GET", "/test", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	test.Assert("Router.ServeHTTP", http.StatusOK, response.Code, t)
	if !handler.IsCalled {
		t.Error("handler of matched route was not called")
	}
}

func TestRouterPost(t *testing.T) {
	router := New()
	handler := NewTestHandler()
	router.Post("/test", handler)
	request, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	test.Assert("Router.ServeHTTP", http.StatusOK, response.Code, t)
	if !handler.IsCalled {
		t.Error("handler of matched route was not called")
	}
}

func TestRouterPut(t *testing.T) {
	router := New()
	handler := NewTestHandler()
	router.Put("/test", handler)
	request, err := http.NewRequest("PUT", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	test.Assert("Router.ServeHTTP", http.StatusOK, response.Code, t)
	if !handler.IsCalled {
		t.Error("handler of matched route was not called")
	}
}

func TestRouterDelete(t *testing.T) {
	router := New()
	handler := NewTestHandler()
	router.Delete("/test", handler)
	request, err := http.NewRequest("DELETE", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	test.Assert("Router.ServeHTTP", http.StatusOK, response.Code, t)
	if !handler.IsCalled {
		t.Error("handler of matched route was not called")
	}
}

func TestRouterServeHTTPNotFound(t *testing.T) {
	router := New()
	request, _ := http.NewRequest("DELETE", "/test", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	test.Assert("Router.ServeHTTP", http.StatusNotFound, response.Code, t)
}

func TestRouteMatch(t *testing.T) {
	route := NewRoute("/test", "GET", NewTestHandler())
	request, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	match := route.Match(request)
	if !match {
		t.Error("Route failed matches request")
	}
}

func TestRouteMatchMethod(t *testing.T) {
	route := NewRoute("/test", "GET", NewTestHandler())
	request, _ := http.NewRequest("POST", "/test", nil)
	match := route.Match(request)
	if match {
		t.Error("Route matches wrong request method")
	}
}

type TestHandler struct {
	IsCalled bool
}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.IsCalled = true
}

func NewTestHandler() *TestHandler {
	return &TestHandler{
		IsCalled: false,
	}
}

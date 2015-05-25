package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testHandler = func(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Hi, Router.")
	res.WriteHeader(http.StatusOK)
}

func TestRouterWithoutParams(t *testing.T) {
	r := NewRouter()
	r.Get("/mosaic/cat", http.HandlerFunc(testHandler))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mosaic/cat", nil)

	r.ServeHTTP(res, req)
	if res.Body.String() != "Hi, Router." {
		t.Errorf("Not equal: %v (expected). %v (actual)", "Hi, Router.", res.Body.String())
	}
}

func TestRouterWithParams(t *testing.T) {
	r := NewRouter()
	r.Get("/mosaic/:id/:size", http.HandlerFunc(testHandler))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mosaic/123456/1024?user=bob", nil)

	r.ServeHTTP(res, req)

	id := req.URL.Query().Get("id")
	if id != "123456" {
		t.Errorf("Not equal: %v (expected). %v (actual)", "123456", id)
	}

	size := req.URL.Query().Get("size")
	if size != "1024" {
		t.Errorf("Not equal: %v (expected). %v (actual)", "1024", id)
	}

	user := req.URL.Query().Get("user")
	if user != "bob" {
		t.Errorf("Not equal: %v (expected). %v (actual)", "bob", id)
	}
}

func TestHeadRequest(t *testing.T) {
	r := NewRouter()
	r.Get("/mosaic/cat", http.HandlerFunc(testHandler))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/mosaic/cat", nil)

	r.ServeHTTP(res, req)
	if res.Body.String() != "Hi, Router." {
		t.Errorf("Not equal: %v (expected). %v (actual)", "Hi, Router.", res.Body.String())
	}
}

func TestRouteNotFound(t *testing.T) {
	r := NewRouter()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)

	r.ServeHTTP(res, req)

	if res.Code != 404 {
		t.Errorf("Not equal: %v (expected). %v (actual)", "404", res.Code)
	}
}

func TestSubrouteNotFound(t *testing.T) {
	r := NewRouter()
	r.Get("/mosaic/:id", http.HandlerFunc(testHandler))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mosaic", nil)

	r.ServeHTTP(res, req)

	if res.Code != 404 {
		t.Errorf("Not equal: %v (expected). %v (actual)", "404", res.Code)
	}
}

func TestSlashendRouteNotFound(t *testing.T) {
	r := NewRouter()
	r.Get("/mosaic/:id", http.HandlerFunc(testHandler))
	r.Get("/", http.HandlerFunc(testHandler))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mosaic/", nil)

	r.ServeHTTP(res, req)

	if res.Code != 404 {
		t.Errorf("Not equal: %v (expected). %v (actual)", "404", res.Code)
	}
}

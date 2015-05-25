package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ptrost/mosaic2go/auth"
	"github.com/ptrost/mosaic2go/config"
	"github.com/ptrost/mosaic2go/image"
	"github.com/ptrost/mosaic2go/model"
	"github.com/ptrost/mosaic2go/test"
)

func TestRESTHandlerServeHTTP(t *testing.T) {
	ctx, _, _ := buildContext()
	response := NewTestResponse()
	isCalled := false
	handler := NewRESTHandler(ctx, func(ctx *Context, r *http.Request) (Response, error) {
		isCalled = true
		return response, nil
	})
	req, _ := http.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if isCalled != true {
		t.Error("RESTHandler.ServeHTTP() didn't call the controller.")
	}
	if response.IsCalled != true {
		t.Error("RESTHandler.ServeHTTP() didn't call response.Write() for the response returned by the controller.")
	}
}

func TestRESTResponseWrite(t *testing.T) {
	response := NewRESTResponse()
	res := httptest.NewRecorder()
	response.Generate(res)
	test.Assert("RESTResponse.Write Header Content-Type", "application/json", res.Header().Get("Content-Type"), t)
	test.Assert("RESTResponse.Write status code", http.StatusOK, res.Code, t)

	var js map[string]interface{}
	err := json.Unmarshal(res.Body.Bytes(), &js)
	test.AssertNotErr("RESTResponse.Write", err, t)
}

func TestHandlerServeHTTP(t *testing.T) {
	ctx, auth, _ := buildContext()
	next := NewTestHandler()
	handler := RequireAuth(ctx, next)
	request, _ := http.NewRequest("GET", "/", nil)
	token := auth.GenerateToken()
	user := model.NewUser(token)
	auth.CreateSession(user)
	request.Header.Add("authToken", token)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	test.Assert("Handler.ServeHTTP", http.StatusOK, response.Code, t)
	if next.IsCalled != true {
		t.Error("Handler.ServeHTTP() didn't call the next handler on successful authentication.")
	}
	if ctx.User == nil {
		t.Error("Handler.ServeHTTP() didn't set the user to the context on successful authentication.")
	}
}

func TestHandlerServeHTTPUnauthorized(t *testing.T) {
	ctx, auth, _ := buildContext()
	next := NewTestHandler()
	handler := RequireAuth(ctx, next)
	request, _ := http.NewRequest("GET", "/", nil)

	token := auth.GenerateToken()
	request.Header.Add("authToken", token)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	test.Assert("Handler.ServeHTTP", http.StatusUnauthorized, response.Code, t)
	if next.IsCalled != false {
		t.Error("Handler.ServeHTTP() called the next handler for unauthorized user.")
	}
}

func TestNewContext(t *testing.T) {
	ctx, _, _ := buildContext()
	_, err := os.Stat(ctx.RootDir)
	test.AssertNotErr("os.Stat(ctx.RootDir)", err, t)
	test.Assert("Context.RootDir", false, os.IsNotExist(err), t)
}

func buildAuth() (*auth.Auth, *config.Config) {
	cfg := config.New(filepath.Join(getRootDir(), "config.json"))
	conn, _ := auth.NewMemcachedConn(cfg)
	auth := auth.New(conn, 1)
	return auth, cfg
}

func buildContext() (*Context, *auth.Auth, *config.Config) {
	auth, config := buildAuth()
	img := image.NewDiscover(StubProvider{})
	ctx := NewContext(config, auth, StubCloudFileStorage{}, img, getRootDir())
	return ctx, auth, config
}

type StubCloudFileStorage struct {
}

func (i StubCloudFileStorage) Save(data io.ReadSeeker, filename string) (string, error) {
	return "", nil
}

func (i StubCloudFileStorage) Delete(filename string) error {
	return nil
}

type StubProvider struct {
}

func (p StubProvider) Search(q string, page int) (*image.SearchResult, error) {
	res := image.SearchResult{}
	res.Images = append(res.Images, "test.jpg")
	return &res, nil
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

type TestResponse struct {
	IsCalled bool
}

func (r *TestResponse) Generate(w http.ResponseWriter) {
	r.IsCalled = true
}

func NewTestResponse() *TestResponse {
	return &TestResponse{
		IsCalled: false,
	}
}

func getRootDir() string {
	_, currentfile, _, _ := runtime.Caller(1)
	return path.Dir(filepath.Join(currentfile, "../"))
}

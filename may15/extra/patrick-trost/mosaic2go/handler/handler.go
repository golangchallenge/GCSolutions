package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"

	"github.com/ptrost/mosaic2go/auth"
	"github.com/ptrost/mosaic2go/config"
	"github.com/ptrost/mosaic2go/image"
	"github.com/ptrost/mosaic2go/model"
	"github.com/ptrost/mosaic2go/router"
)

// RESTHandler forwards the request to a controller that holds the business logic
// of the application. It implements the http.Handler interface.
type RESTHandler struct {
	ctx  *Context
	ctrl Controller
}

// Controller defines a function type that contains the businiess logis of the application
// and is executed by the RESTHandler.
type Controller func(ctx *Context, r *http.Request) (Response, error)

// Response defines an interface that contains all data of the response and the logic
// to generate the output with the http.ResponseWriter. It is returned by the controller.
type Response interface {
	Generate(w http.ResponseWriter)
}

// ServeHTTP passes the http.Request and context to a controller function and writes
// the response to the http.ResponseWriter.
func (h *RESTHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response, err := h.ctrl(h.ctx, r)
	if err != nil {
		log.Println("ERROR: ", err)
		errRes := NewRESTResponse()
		errRes.StatusCode = http.StatusInternalServerError
		errRes.Message = "Something unexpected happened."
		errRes.Generate(w)
	} else {
		response.Generate(w)
	}
}

// NewRESTHandler create an new RESTHandler pointer.
func NewRESTHandler(ctx *Context, ctrl Controller) *RESTHandler {
	return &RESTHandler{ctx, ctrl}
}

// RESTResponse defines the base response data of a REST response, and the logic
// to create a JSON response
type RESTResponse struct {
	Body       map[string]interface{} `json:"body"`
	StatusCode int                    `json:"statusCode"`
	Message    string                 `json:"message"`
}

// Generate creates the JSON response.
func (r *RESTResponse) Generate(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.StatusCode)
	encoder := json.NewEncoder(w)
	err := encoder.Encode(r)
	if err != nil {
		panic(err.Error())
	}
}

// Context holds all global application objects.
type Context struct {
	Config           *config.Config
	Auth             *auth.Auth
	User             *model.User
	CloudFileStorage image.CloudFileStorage
	ImageDiscover    *image.Discover
	RootDir          string
}

// NewContext creates a new Context pointer.
func NewContext(config *config.Config, auth *auth.Auth, storage image.CloudFileStorage, img *image.Discover, rootDir string) *Context {
	return &Context{
		Config:           config,
		Auth:             auth,
		CloudFileStorage: storage,
		ImageDiscover:    img,
		RootDir:          rootDir,
	}
}

// RESTDocHandler generates a documentaion for the REST API. It implements the
// http.Handler interface.
type RESTDocHandler struct {
	ctx    *Context
	routes []*router.Route
}

// ServeHTTP generates a HTML documentation for all given routes and RESTHandler
func (h *RESTDocHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, r)
}

// NewRESTDocHandler create an new RESTDocHandler pointer.
func NewRESTDocHandler(ctx *Context, routes []*router.Route) *RESTDocHandler {
	return &RESTDocHandler{ctx, routes}
}

// AuthHandler implements the http.Handler interface and is responsible for user authentication.
type AuthHandler struct {
	ctx  *Context
	next http.Handler
}

// ServeHTTP reads the an "authToken" header and looks for an active user session for that token.
// If a user is found, the wrapped handler is called and the found user is stored in the context.
// If the request is missing the "authToken" header or no user is found, the handler returns a 401 status code.
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("authToken")
	user, _ := h.ctx.Auth.GetSession(token)

	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		h.ctx.User = user
		_ = h.ctx.Auth.ExtendSession(user)
		h.next.ServeHTTP(w, r)
	}
}

// RequireAuth creates a new Handler with a given context and handler to be
// called when the authentication succeeds.
func RequireAuth(ctx *Context, handler http.Handler) *AuthHandler {
	return &AuthHandler{
		ctx:  ctx,
		next: handler,
	}
}

// NewRESTResponse creates a new NewRESTResponse pointer and sets 200 as default HTTP status code
func NewRESTResponse() *RESTResponse {
	return &RESTResponse{
		StatusCode: http.StatusOK,
		Body:       make(map[string]interface{}),
	}
}

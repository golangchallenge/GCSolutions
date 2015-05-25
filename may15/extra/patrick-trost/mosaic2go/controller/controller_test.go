package controller

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ptrost/mosaic2go/auth"
	"github.com/ptrost/mosaic2go/config"
	"github.com/ptrost/mosaic2go/handler"
	"github.com/ptrost/mosaic2go/image"
	"github.com/ptrost/mosaic2go/model"
	"github.com/ptrost/mosaic2go/test"
)

func TestPostTarget(t *testing.T) {
	_, ctx, _, _ := buildSession()

	req, err := newFileUploadRequest("image", "../test_fixtures/test.jpg")
	test.AssertNotErr("newFileUploadRequest", err, t)
	res, err := PostTarget(ctx, req)
	test.AssertNotErr("PostTarget", err, t)
	restRes := res.(*handler.RESTResponse)

	test.Assert("PostTarget", http.StatusOK, restRes.StatusCode, t)

	expected := ""
	test.Assert("PostTarget response message", expected, restRes.Message, t)
}

func TestPostTargetInvalidMIMEType(t *testing.T) {
	_, ctx, _, _ := buildSession()

	req, err := newFileUploadRequest("image", "../test_fixtures/test.png")
	test.AssertNotErr("newFileUploadRequest", err, t)
	res, err := PostTarget(ctx, req)
	test.AssertNotErr("PostTarget", err, t)

	restRes := res.(*handler.RESTResponse)
	test.Assert("PostTarget response status", http.StatusBadRequest, restRes.StatusCode, t)

	expected := "Invalid file type image/png, only files of type JEPG are allowed."
	test.Assert("PostTarget response message", expected, restRes.Message, t)
}

func TestPostTargetExceedMaxUploadFilesize(t *testing.T) {
	_, ctx, _, _ := buildSession()
	ctx.Config.Set("max_upload_file_size", "1")

	req, err := newFileUploadRequest("image", "../test_fixtures/test.png")
	test.AssertNotErr("newFileUploadRequest", err, t)
	res, err := PostTarget(ctx, req)
	test.AssertNotErr("PostTarget", err, t)

	restRes := res.(*handler.RESTResponse)
	test.Assert("PostTarget response status", http.StatusBadRequest, restRes.StatusCode, t)

	expected := "File is too large. Max upload file size is 0 MB"
	test.Assert("PostTarget response message", expected, restRes.Message, t)
}

func TestGetTarget(t *testing.T) {
	_, ctx, _, _ := buildSession()

	ctx.User.Target = model.NewImage("image.jpg")
	req, err := http.NewRequest("GET", "/", nil)
	res, err := GetTarget(ctx, req)
	test.AssertNotErr("GetTarget", err, t)
	restRes := res.(*handler.RESTResponse)
	test.Assert("GetTarget response status", http.StatusOK, restRes.StatusCode, t)

	image := restRes.Body["image"].(*model.Image)
	test.AssertNot("GetTarget response image.Path", "", image.Path, t)
}

func TestGetTiles(t *testing.T) {
	_, ctx, _, _ := buildSession()

	req, err := http.NewRequest("GET", "/", nil)
	res, err := GetTiles(ctx, req)
	test.AssertNotErr("GetTiles", err, t)
	restRes := res.(*handler.RESTResponse)
	test.Assert("GetTiles response status", http.StatusOK, restRes.StatusCode, t)
}

func newFileUploadRequest(field, file string) (*http.Request, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	defer w.Close()

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	fw, err := w.CreateFormFile(field, file)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(fw, f); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "/", &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
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

func buildAuth() (*auth.Auth, *config.Config) {
	cfg := config.New(filepath.Join(getRootDir(), "config.json"))
	conn, _ := auth.NewMemcachedConn(cfg)
	auth := auth.New(conn, 1)
	return auth, cfg
}

func buildContext() (*handler.Context, *auth.Auth, *config.Config) {
	auth, cfg := buildAuth()
	img := image.NewDiscover(StubProvider{})
	ctx := handler.NewContext(cfg, auth, StubCloudFileStorage{}, img, getRootDir())
	return ctx, auth, cfg
}

func buildSession() (*model.User, *handler.Context, *auth.Auth, *config.Config) {
	ctx, auth, cfg := buildContext()
	token := auth.GenerateToken()
	user := model.NewAnonymousUser(token)
	err := auth.CreateSession(user)
	if err != nil {
		panic(fmt.Sprintf("auth.CreateSession() failed with error: \"%s\".", err.Error()))
	}
	ctx.User = user
	return user, ctx, auth, cfg
}

func getRootDir() string {
	_, currentfile, _, _ := runtime.Caller(1)
	return path.Dir(filepath.Join(currentfile, "../"))
}

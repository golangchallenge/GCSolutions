package controller

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/ptrost/mosaic2go/handler"
	img "github.com/ptrost/mosaic2go/image"
	"github.com/ptrost/mosaic2go/model"
	"github.com/ptrost/mosaic2go/mosaic"
)

// PostTarget uploads a new target image.
func PostTarget(ctx *handler.Context, r *http.Request) (handler.Response, error) {
	response := handler.NewRESTResponse()
	max, _ := strconv.Atoi(ctx.Config.Get("max_upload_file_size"))
	if r.ContentLength > int64(max) {
		response.StatusCode = http.StatusBadRequest
		response.Message = fmt.Sprintf("File is too large. Max upload file size is %v MB", (max / 1024 / 1024))

		return response, nil
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.StatusCode = http.StatusBadRequest
		response.Message = "Missing image in form data."

		return response, nil
	}

	acceptedFileTypes := regexp.MustCompile("image/p?jpeg")
	mimeType := mime.TypeByExtension(path.Ext(header.Filename))
	if !acceptedFileTypes.MatchString(mimeType) {
		response.StatusCode = http.StatusBadRequest
		response.Message = fmt.Sprintf("Invalid file type %s, only files of type JEPG are allowed.", mimeType)

		return response, nil
	}

	filename := path.Join(ctx.User.ID, "target"+path.Ext(header.Filename))
	src, err2 := ctx.CloudFileStorage.Save(file, filename)
	if err2 != nil {
		return nil, err2
	}

	image := model.NewImage(src)
	ctx.User.Target = image
	ctx.Auth.UpdateSession(ctx.User)
	response.Body["image"] = image

	return response, nil
}

// GetTarget returns the users current target image.
func GetTarget(ctx *handler.Context, r *http.Request) (handler.Response, error) {
	response := handler.NewRESTResponse()
	if ctx.User.Target != nil {
		response.Body["image"] = ctx.User.Target
	}
	return response, nil
}

// UserLogin is not implemented yet.
func UserLogin(ctx *handler.Context, r *http.Request) (handler.Response, error) {
	response := handler.NewRESTResponse()

	return response, nil
}

// UserLogout is not implemented yet.
func UserLogout(ctx *handler.Context, r *http.Request) (handler.Response, error) {
	response := handler.NewRESTResponse()

	return response, nil
}

// UserInfo returns the currently logged in user.
func UserInfo(ctx *handler.Context, r *http.Request) (handler.Response, error) {
	response := handler.NewRESTResponse()
	response.Body["user"] = ctx.User

	return response, nil
}

// UserAnonymous creates an new anonymous user and auth token.
func UserAnonymous(ctx *handler.Context, r *http.Request) (handler.Response, error) {
	response := handler.NewRESTResponse()
	token := ctx.Auth.GenerateToken()
	user := model.NewAnonymousUser(token)
	err := ctx.Auth.CreateSession(user)
	if err != nil {
		return nil, err
	}
	response.Body["user"] = user
	response.Body["authToken"] = token

	return response, nil
}

// GetTiles searches tile images
func GetTiles(ctx *handler.Context, r *http.Request) (handler.Response, error) {
	response := handler.NewRESTResponse()
	q := r.URL.Query().Get("q")
	if q == "" {
		q = "cats"
	}
	var images []*model.Image
	result, err := ctx.ImageDiscover.Search(q, 1)
	if err != nil {
		return nil, err
	}
	for _, imgPath := range result.Images {
		images = append(images, model.NewImage(imgPath))
	}

	response.Body["images"] = images
	return response, nil
}

// PostMosaic creates a Mosaic for the current target image and tiles that matches a specified query
func PostMosaic(ctx *handler.Context, r *http.Request) (handler.Response, error) {
	response := handler.NewRESTResponse()
	if ctx.User.Target == nil {
		response.Body["message"] = "Upload a target image first."
		return response, nil
	}

	// Get Tiles for the given query
	q := r.URL.Query().Get("q")
	if q == "" {
		q = "cats"
	}
	result, err := ctx.ImageDiscover.Search(q, 1)
	if err != nil {
		return nil, err
	}

	// Download tiles to the  a directory with a name generated from the tile search query
	tileDir := path.Join(ctx.RootDir, ctx.Config.Get("tmp_dir"), img.GenerateDirname(q))
	errMkdir := os.MkdirAll(tileDir, 0700)
	if errMkdir != nil {
		return nil, err
	}

	// Download tile images
	errDwnl := ctx.ImageDiscover.Fetch(result.Images, tileDir)
	if errDwnl != nil {
		return nil, errDwnl
	}

	// Create the tile image pool
	tileFiles, _ := ioutil.ReadDir(tileDir)
	var pool []image.Image
	for _, f := range tileFiles {
		poolImg, err := img.Open(filepath.Join(tileDir, f.Name()))
		if err != nil {
			return nil, err
		}
		pool = append(pool, poolImg)
	}

	if len(pool) < 100 {
		response.StatusCode = http.StatusBadRequest
		response.Body["message"] = "Not enough tile images to generate the mosaic."
		return response, nil
	}

	targetFilename := fmt.Sprintf("%s-target.jpg", ctx.User.ID)
	targetSrc := filepath.Join(ctx.RootDir, ctx.Config.Get("tmp_dir"), targetFilename)

	// Download the target to the tmp dir
	_, errTargetDwnl := ctx.ImageDiscover.DownloadFile(ctx.User.Target.Path, targetSrc)
	if errTargetDwnl != nil {
		return nil, errTargetDwnl
	}

	targetImg, errOpenTarget := img.Open(targetSrc)
	if errOpenTarget != nil {
		return nil, errOpenTarget
	}

	// Generate the mosaic
	m := mosaic.New(targetImg)
	mosaicImg, _ := m.Generate(pool)

	var buf bytes.Buffer
	errDecode := jpeg.Encode(&buf, mosaicImg, &jpeg.Options{Quality: jpeg.DefaultQuality})
	if errDecode != nil {
		return nil, errDecode
	}

	// Upload the mosaic to S3
	filename := path.Join(ctx.User.ID, "mosaic.jpg")
	mosaicSrc, errStore := ctx.CloudFileStorage.Save(bytes.NewReader(buf.Bytes()), filename)
	if errStore != nil {
		return nil, errStore
	}

	image := model.NewImage(mosaicSrc)
	ctx.User.Mosaic = image
	ctx.Auth.UpdateSession(ctx.User)
	response.Body["image"] = image

	return response, nil
}

// GetMosaic returns the users current mosaic image.
func GetMosaic(ctx *handler.Context, r *http.Request) (handler.Response, error) {
	response := handler.NewRESTResponse()
	if ctx.User.Mosaic != nil {
		response.Body["image"] = ctx.User.Mosaic
	}
	return response, nil
}

package main

import (
	"log"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/ptrost/mosaic2go/auth"
	"github.com/ptrost/mosaic2go/config"
	"github.com/ptrost/mosaic2go/controller"
	"github.com/ptrost/mosaic2go/handler"
	"github.com/ptrost/mosaic2go/image"
	"github.com/ptrost/mosaic2go/router"
)

func main() {
	router := router.New()
	rootDir := getRootDir()
	cfg := config.New(filepath.Join(rootDir, "config.json"))
	conn, _ := auth.NewMemcachedConn(cfg)
	ttl, _ := strconv.Atoi(cfg.Get("session_expire"))
	ath := auth.New(conn, ttl)
	storage := image.NewS3FileStorage(cfg.Get("s3_region"), cfg.Get("s3_images_bucket"))
	imgProvider := image.NewFlickrProvider(cfg.Get("flickr_key"), cfg.Get("flickr_secret_key"))
	img := image.NewDiscover(imgProvider)
	ctx := handler.NewContext(cfg, ath, storage, img, rootDir)

	router.Post("/user/anonymous", handler.NewRESTHandler(ctx, controller.UserAnonymous))
	router.Get("/user/info", handler.RequireAuth(ctx, handler.NewRESTHandler(ctx, controller.UserInfo)))
	router.Get("/mosaic", handler.RequireAuth(ctx, handler.NewRESTHandler(ctx, controller.GetMosaic)))
	router.Post("/mosaic", handler.RequireAuth(ctx, handler.NewRESTHandler(ctx, controller.PostMosaic)))
	router.Post("/mosaic/target", handler.RequireAuth(ctx, handler.NewRESTHandler(ctx, controller.PostTarget)))
	router.Get("/mosaic/target", handler.RequireAuth(ctx, handler.NewRESTHandler(ctx, controller.GetTarget)))
	router.Get("/mosaic/tiles", handler.RequireAuth(ctx, handler.NewRESTHandler(ctx, controller.GetTiles)))
	router.Get("/", handler.NewRESTDocHandler(ctx, router.Routes))

	http.Handle("/", router)

	host := ":" + cfg.Get("PORT")
	log.Println("Starting web server on", host)
	if err := http.ListenAndServe(host, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func getRootDir() string {
	_, currentfile, _, _ := runtime.Caller(1)
	return path.Dir(currentfile)
}

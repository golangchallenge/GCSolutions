package main

import (
	"database/sql"
	"github.com/aishraj/gopherlisa/app"
	"github.com/aishraj/gopherlisa/common"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
)

func init() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}

func main() {
	sessionStore, err := common.NewSessionManager("gopherId", 3600)
	if err != nil {
		log.Fatal("Unable to start the session store manager.", err)
	}

	Info := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info.Println("Starting out the go program")
	db, err := sql.Open("mysql", common.MySqlUserName+":"+common.MySqlPassword+"@/gopherlisa") //TODO Read this from environment or config
	if err != nil {
		log.Fatal("Unable to get a connection to MySQL. Error is: ", err)
	}

	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Images ( id int(5) NOT NULL AUTO_INCREMENT, imgtype varchar(255),  img varchar(255), red int(16), green int(16), blue int(16), PRIMARY KEY(id) )")
	if err != nil {
		log.Fatal("Unable to create table in db. Aborting now. Error is :", err)
	}
	context := &common.AppContext{Log: Info, SessionStore: sessionStore, Db: db}
	context.Log.Println("Server starting on port: ", common.LocalHttpPort)
	authHandler := app.Handler{AppContext: context, HandlerFunc: app.AuthroizeHandler}
	rootHandler := app.Handler{AppContext: context, HandlerFunc: app.BaseHandler}
	uploadHandler := app.Handler{AppContext: context, HandlerFunc: app.UploadHandler}
	searchHandler := app.Handler{AppContext: context, HandlerFunc: app.SearchHandler}
	http.Handle("/login/", authHandler)
	http.Handle("/search", searchHandler)
	http.Handle("/upload/", uploadHandler)
	http.Handle("/", rootHandler)
	http.ListenAndServe(":"+common.LocalHttpPort, nil)
}

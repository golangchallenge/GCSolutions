package imgtools

import (
	"database/sql"
	"github.com/aishraj/gopherlisa/common"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"testing"
)

func TestIndex(t *testing.T) {
	t.Skip("skipping test mode.") //Again, this test won't work until you have a real db setup. TODO: Use a mock db.
	sessionStore, err := common.NewSessionManager("gopherId", 7200)
	if err != nil {
		log.Fatal("Unable to start the session store manager.", err)
	}
	Info := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info.Println("Starting out the go program")
	db, err := sql.Open("password", "root:mysql@/gopherlisa")
	if err != nil {
		log.Fatal("Unable to get a connection to MySQL. Error is: ", err)
	}

	context := &common.AppContext{Info, sessionStore, db} //TODO: add db connection
	directoryName := "kathmandu"                          //TODO: screw it  for now i'm going to remove this test and revoke access later
	n, err := AddImagesToIndex(context, directoryName)
	if err != nil {
		log.Fatal("ERROR!!!!", err)
	}
	log.Println("The number of images we got are", n)
}

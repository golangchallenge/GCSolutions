package imgtools

import (
	"database/sql"
	"github.com/aishraj/gopherlisa/common"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"testing"
)

func TestMatcher(t *testing.T) {
	t.Skip("skipping test mode.") //This test won't workuntil you got the file and the data in the db
	sessionStore, err := common.NewSessionManager("gopherId", 7200)
	if err != nil {
		log.Fatal("Unable to start the session store manager.", err)
	}
	Info := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info.Println("Starting out the go program")
	db, err := sql.Open("mysql", "root:mysql@/gopherlisa")
	if err != nil {
		log.Fatal("Unable to get a connection to MySQL. Error is: ", err)
	}

	context := &common.AppContext{Info, sessionStore, db} //TODO: add db connection
	loadedImage, err := LoadFromDisk(common.DownloadBasePath + "kathmandu/10684246_1100050490011860_943310652_n.jpg")
	if err != nil {
		context.Log.Fatal("Cannot load image from disk.", err)
	}
	matchedImage := FindProminentColour(loadedImage)
	nearestImage := findClosestMatch(context, matchedImage, "kathmandu")
	context.Log.Println("neareset one is", nearestImage)

}

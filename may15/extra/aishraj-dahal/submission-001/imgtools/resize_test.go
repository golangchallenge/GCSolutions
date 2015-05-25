package imgtools

import (
	"github.com/aishraj/gopherlisa/common"
	"log"
	"os"
	"testing"
)

func TestResize(t *testing.T) {
	t.Skip("skipping test mode.") //This test could work. But it would require some work around for the path.
	sessionStore, err := common.NewSessionManager("gopherId", 3600)
	if err != nil {
		log.Fatal("Unable to start the session store manager.", err)
	}
	Info := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info.Println("Starting out the go program")
	context := &common.AppContext{Info, sessionStore, nil} //TODO: add db connection
	directoryName := "kathmandu"                           //TODO: screw it  for now i'm going to remove this test and revoke access later
	n, ok := ResizeImages(context, directoryName)
	if !ok {
		log.Println("Unable to read images.")
	}
	log.Println("The number of images we got are", n)
}

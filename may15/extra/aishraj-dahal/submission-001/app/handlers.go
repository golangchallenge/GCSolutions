package app

import (
	"errors"
	"fmt"
	"github.com/aishraj/gopherlisa/common"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Handler struct {
	AppContext  *common.AppContext
	HandlerFunc func(*common.AppContext, http.ResponseWriter, *http.Request) (int, error)
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := handler.HandlerFunc(handler.AppContext, w, r)
	if err != nil {
		log.Printf("HTTP %d: %q", status, err)
		switch status {
		case http.StatusNotFound:
			http.NotFound(w, r)
		case http.StatusInternalServerError:
			http.Error(w, http.StatusText(status), status)
		default:
			http.Error(w, http.StatusText(status), status)
		}
	}
	if status == http.StatusSeeOther {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func validateAndStartSession(context *common.AppContext, w http.ResponseWriter, r *http.Request) common.Session {
	session := context.SessionStore.SessionStart(w, r)
	createtime := session.Get("createtime")
	if createtime == nil {
		session.Set("createtime", time.Now().Unix())
	} else if (createtime.(int64) + 360) < (time.Now().Unix()) {
		context.Log.Println("Session has expired, starting new session.")
		context.SessionStore.SessionDestroy(session.SessionID())
		session = context.SessionStore.SessionStart(w, r)
	}
	return session
}

func BaseHandler(context *common.AppContext, w http.ResponseWriter, r *http.Request) (retVal int, err error) {
	session := validateAndStartSession(context, w, r)
	switch r.Method {
	case "GET":
		code := r.URL.Query().Get("code")
		instaError := r.URL.Query().Get("error")
		if instaError == "access_denied" && r.URL.Query().Get("error_reason") == "user_denied" {
			context.Log.Println("User denied permission for access.")
			retVal = http.StatusUnauthorized
			err = errors.New("Seems you didn't allow that to happen")
			return
		}
		if len(code) != 0 {
			//This is a callback from instagram to get the token. We now call a method that retuns us either the token or the error message.
			authToken, er := GetAuthToken(context, w, r, code)
			if er != nil {
				context.Log.Println("Unable to get the token. Error was: ", er)
				return http.StatusInternalServerError, errors.New(er.Error())
			}
			session := context.SessionStore.SessionStart(w, r)
			session.Set("user", authToken.User.FullName)
			session.Set("access_token", authToken.AccessToken)
			session.Set("userId", authToken.User.ID)
			//Now redirect the user to the upload page. (ie redirect to the hoemage again, but display the upload template instead)
			return http.StatusSeeOther, nil
		}
		//now that everything's done, we try to render the right template
		// ie upload template if the session has the user token but not the uploaded flag, search if both the user token and uploaded flag, else the root
		//lets read the session token
		displayUser := session.Get("user")
		uploadedFlag := session.Get("uploaded")

		markup := renderIndex(context, displayUser, uploadedFlag)
		if markup == nil {
			context.Log.Println("Unable to render the templates.")
			return http.StatusInternalServerError, nil
		}
		fmt.Fprint(w, string(markup))
		context.Log.Println("Done generating markup.")
		return http.StatusOK, nil
	default:
		return http.StatusUnauthorized, nil
	}
}

func UploadHandler(context *common.AppContext, w http.ResponseWriter, r *http.Request) (revVal int, err error) {
	session := context.SessionStore.SessionStart(w, r)
	authToken := session.Get("access_token")
	if authToken == nil {
		return http.StatusUnauthorized, errors.New("Cannot authorize")
	}
	userId := session.Get("userId")
	if userId == nil {
		return http.StatusInternalServerError, errors.New("UserId not there in sesion. ERROR")
	}
	fileId, ok := userId.(string)
	if !ok {
		context.Log.Println("Unable to cast the userid from session storage.")
		return http.StatusInternalServerError, errors.New("Cannot cast the user id from session storage.")
	}
	switch r.Method {
	case "GET":
		context.Log.Print("Method is get - attempting to render the upload template")
		markup := executeTemplate(context, "upload", nil)
		if markup == nil {
			context.Log.Println("Unable to render the upload template")
			return http.StatusInternalServerError, errors.New("Unable to render the upload template")
		}
		params := map[string]interface{}{"LayoutContent": template.HTML(string(markup))}
		pageMarkup := executeTemplate(context, "head", params)
		fmt.Fprint(w, string(pageMarkup))
		return http.StatusOK, nil
	case "POST":
		context.Log.Println("The method is post, now trying to get the input file.")
		file, header, err := r.FormFile("file")

		if err != nil {
			context.Log.Println("Could not upload the file ******", err)
			return http.StatusInternalServerError, err
		}

		defer file.Close()
		context.Log.Println("Creting the file in local file system")
		out, err := os.Create("/tmp/" + fileId + ".jpg")
		if err != nil {
			context.Log.Println("Unable to create the file for writing. Check your write access privilege")
			return http.StatusInternalServerError, err
		}
		defer out.Close()
		context.Log.Println("Populating local file data")
		_, err = io.Copy(out, file)
		if err != nil {
			context.Log.Println(w, err)
		}

		context.Log.Println("File uploaded successfully : ", header.Filename)

		err = session.Set("uploaded", true)
		if err != nil {
			context.Log.Fatal("Unable to set the value in the session. Error is:", err)
		}

		//now that our images are in the index, display the image upload page
		context.Log.Println("Now redirecting to the index handler")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		revVal = http.StatusOK
		err = nil
		return revVal, err
	}
	return http.StatusMethodNotAllowed, errors.New("This method is not allowed on this resource.")

}

func renderIndex(context *common.AppContext, userWrapper, uploadedFlag interface{}) []byte {
	// Generate the markup for the index template.
	if userWrapper == nil {
		context.Log.Print("Attempting to render the login template")
		markup := executeTemplate(context, "login", nil)
		if markup == nil {
			return nil
		}
		params := map[string]interface{}{"LayoutContent": template.HTML(string(markup))}
		return executeTemplate(context, "head", params)
	} else if uploadedFlag == nil {
		context.Log.Print("Attempting to render the upload template")
		markup := executeTemplate(context, "upload", nil)
		if markup == nil {
			return nil
		}
		params := map[string]interface{}{"LayoutContent": template.HTML(string(markup))}
		return executeTemplate(context, "head", params)
	}
	context.Log.Print("Attempting to render the search template")
	markup := executeTemplate(context, "search", nil)
	if markup == nil {
		return nil
	}
	params := map[string]interface{}{"LayoutContent": template.HTML(string(markup))}
	return executeTemplate(context, "head", params)
}

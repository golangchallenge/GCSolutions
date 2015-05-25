package app

import (
	"bytes"
	"errors"
	"github.com/aishraj/gopherlisa/common"
	"github.com/aishraj/gopherlisa/imgtools"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func SearchHandler(context *common.AppContext, w http.ResponseWriter, r *http.Request) (revVal int, err error) {
	if r.Method == "POST" {
		session := context.SessionStore.SessionStart(w, r)
		authToken := session.Get("access_token")
		context.Log.Println("Already authenticated, and therefore fetching the image from instagram now.")
		formData := r.FormValue("searchTerm")
		tokenString, ok := authToken.(string)
		if !ok {
			context.Log.Println("Token cannot be cast to string. ERROR.")
			return http.StatusInternalServerError, errors.New("Token cannot be cast to string")
		}
		context.Log.Println("^^^^^^ The search query is ^^^^^", formData)
		if len(formData) < 1 {
			context.Log.Fatal("Cannot continue, the form data cannot have a length less than 1")
		}
		//small trick, if the directory for the searchterm Already has the required number of files just skip
		if isDownloadRequired(context, formData) == true {
			context.Log.Println("Download is required, downloading now")
			//TODO: if the files are fewer than download and index. (nice to have)
			if !directoryExists(context, formData) {
				err := os.Mkdir(common.DownloadBasePath+formData, 0777)
				if err != nil {
					context.Log.Println("not able to create the directory ***************")
				}
			}

			images, err := LoadImages(context, formData, tokenString)
			if err != nil {
				context.Log.Println("Error fetching from instagram.")
				return http.StatusInternalServerError, err
			}
			context.Log.Println("List of Images we got are:", images)
			downloadCount, ok := DownloadImages(context, images, formData)
			if !ok {
				context.Log.Println("Unable to download images to the path")
				return http.StatusInternalServerError, errors.New("Download failed")
			}
			context.Log.Println("Download count was: ", downloadCount)
			n, ok := imgtools.ResizeImages(context, formData)
			if !ok {
				context.Log.Println("Unable to resize images")
				return http.StatusInternalServerError, errors.New("Resizing images failed")
			}
			context.Log.Println("Number of images resized was ", n)

			n, err = imgtools.AddImagesToIndex(context, formData)
			if err != nil {
				context.Log.Println("Unable to add images to index", err)
				return http.StatusInternalServerError, err
			}
			context.Log.Println("Number of images indexed was", n)
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

		generatedImage := imgtools.CreateMosaic(context, fileId, formData)
		//image TODO get resize working first

		buffer := new(bytes.Buffer)
		if err := jpeg.Encode(buffer, generatedImage, nil); err != nil {
			context.Log.Println("unable to encode image.")
		}

		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
		if _, err := w.Write(buffer.Bytes()); err != nil {
			context.Log.Println("unable to write image.")
		}

		return http.StatusOK, nil
	}
	return http.StatusMethodNotAllowed, errors.New("This method is not allowed here.")

}

func isDownloadRequired(context *common.AppContext, searchTerm string) bool {
	files, err := ioutil.ReadDir(common.DownloadBasePath + searchTerm)
	if err != nil {
		context.Log.Println("ERROR: Unable to count the number of files")
		return true //yes download
	}
	fileCount := len(files)
	context.Log.Println("The number of files in the tag dir is", fileCount)
	if fileCount > 1000 { //TODO change it to less than when testing is over
		return true
	}
	return false
}

func directoryExists(context *common.AppContext, dirname string) bool {
	src, err := os.Stat(common.DownloadBasePath + dirname)
	if err != nil {
		context.Log.Println("Unable to verify OS stat.")
		return false
	}

	// check if the source is indeed a directory or not
	if !src.IsDir() {
		return false
	}
	return true
}

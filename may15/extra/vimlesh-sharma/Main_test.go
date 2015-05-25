package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestMossaicForPixelWidth5_test(t *testing.T) {
	FormData := make(map[string]string)
	FormData["searchtext"] = "Blue Sky"
	FormData["SelectedTileSize"] = "5"
	response, err := PostData("http://localhost:8000/upload", FormData, "./TestImage/Sky.jpg")
	if err != nil {
		t.Error(err)
	}
	FormData = make(map[string]string)
	FormData["PUID"] = GetValueFor("PUID", response)
	FormData["TileSize"] = GetValueFor("TileSize", response)
	FormData["ImageName"] = GetValueFor("ImageName", response)
	_, err = PostData("http://localhost:8000/mossaic", FormData, "")
	if err != nil {
		t.Error(err)
	}
}

func TestMossaicForPixelWidth20AndFruits_test(t *testing.T) {
	FormData := make(map[string]string)
	FormData["searchtext"] = "Fruits"
	FormData["SelectedTileSize"] = "20"
	response, err := PostData("http://localhost:8000/upload", FormData, "./TestImage/Apples.jpg")
	if err != nil {
		t.Error(err)
	}
	FormData = make(map[string]string)
	FormData["PUID"] = GetValueFor("PUID", response)
	FormData["TileSize"] = GetValueFor("TileSize", response)
	FormData["ImageName"] = GetValueFor("ImageName", response)
	_, err = PostData("http://localhost:8000/mossaic", FormData, "")
	if err != nil {
		t.Error(err)
	}
}

func PostData(url string, FormData map[string]string, uploadImageName string) (string, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	if uploadImageName != "" {
		err := WriteFormFile(bodyWriter, uploadImageName)
		if err != nil {
			return "", err
		}
	}
	if len(FormData) > 0 {
		err := WriteFormData(bodyWriter, FormData)
		if err != nil {
			return "", err
		}
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	response, err := PostRequest(url, contentType, bodyBuf)
	if err != nil {
		return "", err
	}

	return response, nil
}

func PostRequest(targetUrl string, contentType string, bodyBuf *bytes.Buffer) (string, error) {
	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodystr := string(resp_body)
	return bodystr, nil
}

func WriteFormData(bw *multipart.Writer, FormData map[string]string) error {
	for key, value := range FormData {
		formFieldWriter, err := bw.CreateFormField(key)
		if err != nil {
			return err
		}
		formFieldWriter.Write([]byte(value))
	}
	return nil
}

func WriteFormFile(bw *multipart.Writer, filename string) error {

	fileWriter, err := bw.CreateFormFile("file", filename)
	if err != nil {
		return err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		return err
	}

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}
	return nil
}

func GetValueFor(Searchtext string, SourceString string) string {
	s1 := SourceString[strings.Index(SourceString, Searchtext):]
	s1 = s1[strings.Index(s1, "value="):]
	s1 = s1[len("value="):]
	s2 := ""

	foundIdx := 0
	for i, j := range s1 {
		foundIdx = i
		if j == '>' {
			break
		}
	}
	s2 = s1[:foundIdx]
	return s2
}

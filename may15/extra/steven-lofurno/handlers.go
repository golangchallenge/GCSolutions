// handlers
package main

import (
	"image/jpeg"
	"io"
	"image"
	"net/http"
	"fmt"
	"strconv"	
	"encoding/json"
	"bytes"
	"io/ioutil"
	"log"
)

//returns a json representation of our saved mosaics
func getImages(w http.ResponseWriter, req *http.Request) {
	
	r,_:=json.Marshal(savedMosaics)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte(r))
}

//accepts source image and relevant parameters and enqueues a mosaic request
func postImage(w http.ResponseWriter, req *http.Request) {
		
	defer req.Body.Close()
	buf, err := ioutil.ReadAll(req.Body)
	buffer:=bytes.NewReader(buf)

	m, _, err := image.Decode(buffer)
	
	if err != nil {
		log.Fatal(err)
		return
	}
	
	rgba := convertImage(m)
	
	qs:=req.URL.Query()
	terms:= qs["terms"]
	tosave:= false
	tosave,err=strconv.ParseBool(qs.Get("save"))
		
	mr := newMosRequest(rgba,terms,tosave)
	mr.Key=randomString(16)
	
	mosRequests[mr.Key] = mr
	
	mosQueue <- mr
	mr.Progress<-"mosaic queued"

	w.Write([]byte(mr.Key))
}

//streams status updates over http, including our final mosaic image
func listen(w http.ResponseWriter, req *http.Request) {
	
	key := req.URL.Query().Get("key")
		
	h, _ := w.(http.Hijacker)
	conn, rw, _ := h.Hijack()
	defer conn.Close()
	
	rw.Write([]byte("HTTP/1.1 200 OK\r\n"))
	rw.Write([]byte("Content-Type: text/event-stream\r\n\r\n"))
	rw.Flush()
	
	var mr *mosRequest
	var ok bool
	
	if mr, ok = mosRequests[key]; !ok {
		fmt.Println("key not found")
    return
	}
	delete(mosRequests,key)
		
	disconnect:=make(chan bool, 1)
	
	go func(){
		_,err := rw.ReadByte()
		if err==io.EOF {
			disconnect<-true
		}
	}()
	
	for {
		
		select{
		case <-disconnect:
			fmt.Println("disconnected")
			return
			
		case msg := <-mr.Progress:
			
			rw.Write([]byte("event: progress\n"))
			rw.Write([]byte("data: " + msg+"\n\n"))
			rw.Flush()
			
		case mosaic := <-mr.Result:
			
			var b bytes.Buffer	
	
			jpeg.Encode(&b,mosaic,nil)
			
			//str := base64.StdEncoding.EncodeToString(result.Mosaic.Bytes())
			//json:= "{\"height\":" + strconv.Itoa(result.Height) + ",\"width\":" + strconv.Itoa(result.Width) + ",\"base64\":\""+str+ "\"}"
			bb,_:= json.Marshal(b.Bytes())		
			
			rw.Write([]byte("event: image\n"))
			rw.Write([]byte("data: "))
			rw.Write(bb)
			//rw.Write([]byte(json))
			rw.Write([]byte("\n\n"))
			rw.Flush()
			return				
		}		
	}		
}
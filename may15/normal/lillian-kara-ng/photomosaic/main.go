package main

import (
	"bitbucket.org/lillian_ng/photomosaic/mosaic"
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/create.html")
}

func mosaicHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	targetUrl := r.PostForm["targeturl"][0]
	hashtag := r.PostForm["hashtag"][0]
	clientId := r.PostForm["instagram-client-id"][0]

	fmt.Fprintln(w, startMosaicResponse)

	f, ok := w.(http.Flusher)
	if ok && f != nil {
		f.Flush()
	}

	fmt.Fprintln(w, mosaic.BuildMosaic(targetUrl, hashtag, clientId),
		finishMosaicResponse)
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/mosaic", mosaicHandler)
	http.ListenAndServe(":8080", nil)

}

var startMosaicResponse = `
<html>
<head>
  <title>Mosaic</title>
</head>

<body>

<div id="loading">Loading...</div>
`
var finishMosaicResponse = `
<script type="text/javascript">
	document.getElementById('loading').style.display = 'none';
</script>

</body>
</html>
`

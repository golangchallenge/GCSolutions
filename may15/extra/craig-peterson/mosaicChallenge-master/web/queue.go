package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/captncraig/mosaicChallenge/imgur"
	"github.com/captncraig/mosaicChallenge/mosaics"
)

type jobStatus struct {
	Status, Substatus string
	Version           int
}

type job struct {
	id         string
	mainImgUrl string
	gallery    string
	status     jobStatus

	token    *imgur.ImgurAccessToken
	l        sync.Mutex
	watchers []chan<- jobStatus
}

func createJob(imgUrl, gallery string, token *imgur.ImgurAccessToken) (id string) {
	id = randSeq(8)
	j := job{
		id:         id,
		mainImgUrl: imgUrl,
		gallery:    gallery,
		status:     jobStatus{},
		l:          sync.Mutex{},
		watchers:   []chan<- jobStatus{},
		token:      token,
	}
	jobMutex.Lock()
	allJobs[id] = &j
	jobMutex.Unlock()
	jobSubmission <- &j
	return id
}

func (j *job) updateQueuePosition(pos int) {
	j.updateStatus("Waiting for an available worker", fmt.Sprintf("position %d in queue", pos))
}

func (j *job) updateStatus(status, substatus string) {
	j.l.Lock()
	j.status.Status = status
	j.status.Substatus = substatus
	j.status.Version++
	for _, watcher := range j.watchers {
		watcher <- j.status
	}
	j.watchers = []chan<- jobStatus{}
	j.l.Unlock()
}

func (j *job) subscribe(ch chan<- jobStatus) {
	j.l.Lock()
	j.watchers = append(j.watchers, ch)
	j.l.Unlock()
}

// All jobs currently in the pipeline
var allJobs = map[string]*job{}
var jobMutex = sync.RWMutex{}

//channel to submit new jobs on
var jobSubmission = make(chan *job)

func runWorkQueue() {
	//channel to give jobs to workers
	var work = make(chan *job)

	log.Printf("Starting %d worker routines.\n", runtime.NumCPU())
	for i := 0; i < 1; i++ {
		go worker(work)
	}

	queue := []*job{}
	enqueue := func(j *job) {
		queue = append(queue, j)
		j.updateQueuePosition(len(queue))
	}
	for {
		// if anything in the queue we send and receive concurrently
		// if queue is empty we only receive
		if len(queue) == 0 {
			select {
			case j := <-jobSubmission:
				enqueue(j)
			}
		} else {
			select {
			case j := <-jobSubmission:
				enqueue(j)
			case work <- queue[0]:
				queue = queue[1:]
				for i, j := range queue {
					j.updateQueuePosition(i + 1)
				}
			}
		}
	}
}

func worker(q <-chan *job) {
	for {
		j := <-q
		j.updateStatus("Prepping data", "downloading source image")
		resp, err := http.Get(j.mainImgUrl)
		if err != nil {
			j.updateStatus("Done", "error downloading image")
			log.Println(err)
			continue
		}
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			j.updateStatus("Done", "error decoding image")
			log.Println(err)
			continue
		}
		j.updateStatus(j.status.Status, "verifying library")
		lib, ok := collections[j.gallery]
		if !ok {
			j.updateStatus("Done", "error: gallery does not exist")
			continue
		}
		j.updateStatus("Building mosaic", "")
		reporter := make(chan float64)
		done := make(chan struct{})
		var mosaic image.Image
		go func() {
			mosaic = mosaics.BuildMosaicFromLibrary(img, lib.library, reporter)
			close(done)
		}()
	Loop:
		for {
			select {
			case pct := <-reporter:
				j.updateStatus("Building mosaic", fmt.Sprintf("%.2f%% percent complete", pct))
			case <-done:
				break Loop
			}
		}
		j.updateStatus("Encoding image", "")
		buf := &bytes.Buffer{}
		jpeg.Encode(buf, mosaic, &jpeg.Options{30})
		j.updateStatus("Done", fmt.Sprintf("<img width=1000px src='data:image/jpg;base64,%s'/><br/>Right click to save.<br/><a href='/options'>Build another!</a>", base64.StdEncoding.EncodeToString(buf.Bytes())))
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

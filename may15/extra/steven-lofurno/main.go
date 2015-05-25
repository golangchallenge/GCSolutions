// main
package main

import (

	"image/jpeg"
	"math/rand"
	"time"
	"runtime"
	"net/http"
	"image"
	"image/png"
	"image/draw"
	"fmt"
	"os"
	_ "image/png"
	_ "image/jpeg"
)

const (
	tileWidth = 8
	tileHeight = 8
	tileXResolution =2
	tileYResolution =2
	mosaicScale=8
	maxColorDifference = 120.0
	alpha = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)



var mosRequests = make(map[string]*mosRequest)
var mosQueue = make(chan *mosRequest, 100)
var savedMosaics []string

func newMosRequest(img *image.RGBA, terms []string, tosave bool) *mosRequest {
	
	r:=&mosRequest{}
	r.Image=img
	r.Terms=terms
	r.Progress = make(chan string, 15)
	r.Result = make(chan *image.RGBA, 1)
	r.Save=tosave
	
	return r
}


func saveJPG(img *image.RGBA, fn string) {
	
	f,err := os.OpenFile(fn,os.O_CREATE|os.O_RDWR, 0666)
	
	
	defer func() { 
    if err := f.Close(); err != nil {
         fmt.Println(err)
    }
	}()
	
	if err != nil{
		fmt.Println(err)
		return
	}
	
	err = jpeg.Encode(f, img,nil)
		if err != nil{
		fmt.Println(err)
		return
	}
	
}

//saves image to disk in png format
func saveImage(img *image.RGBA, fn string) {
	
	f,err := os.OpenFile(fn,os.O_CREATE|os.O_RDWR, 0666)
		
	defer func() { 
    if err := f.Close(); err != nil {
         fmt.Println(err)
    }
	}()
	
	if err != nil{
		fmt.Println(err)
		return
	}
	
	err = png.Encode(f, img)
		if err != nil{
		fmt.Println(err)
		return
	}	
}

//true about 50% of the time
func coinFlip() bool {
	
	if rand.Float64()>=.5 {
		return true
	}
	return false
	
}



//returns a string of length l, with chars randomly picked from the following
//chars: abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789
func randomString(l int ) string {
	 
    bytes := make([]byte, l)
    for i:=0 ; i<l ; i++ {
			bytes[i]= alpha[rand.Intn(len(alpha))]
    }
    return string(bytes)
}


func init(){
	rand.Seed( time.Now().UTC().UnixNano())
			
	f,err := os.Open("static/images")//File("static/images",os.O_CREATE, 0666)
	defer f.Close()
	
	if err!=nil {
		return
	}
	
	fi,err:=f.Readdir(200)
	
	if err!=nil {
		return
	}
	
	for _,file:=range fi{
		savedMosaics=append(savedMosaics,file.Name())
	}
}

//downloads images from urls in source queue, proceses them, and streams results
func imageDownloader(queue <-chan string, results chan<- imageResponse) {
    for q := range queue {
			m,err := downloadAndDecode(q)
      results <- imageResponse{Image:m,Err:err}
    }
}

//composes a mosaic from image tiles, to match a provided source RGBA image
func fitMosaic(rgba *image.RGBA, tiles []mosImage) *image.RGBA {
	
	var dx = tileXResolution
	var dy = tileYResolution
	
	height:=rgba.Bounds().Max.Y
	width:=rgba.Bounds().Max.X
	
	outscalex:=tileWidth/tileXResolution
	outscaley:=tileHeight/tileYResolution
	
	downx:=width/outscalex
	downy:=height/outscaley
	
	mosaicx:=downx*mosaicScale*(tileWidth/tileXResolution)
	mosaicy:=downy*mosaicScale*(tileHeight/tileYResolution)
	
	out:=downsample(rgba, image.Rect(0,0,downx,downy))
	mosaic := image.NewRGBA(image.Rect(0,0,mosaicx,mosaicy))
	
	const tileScale = tileHeight*mosaicScale
		
	for j:=0;j<out.Bounds().Max.Y;j+=dy {
		for i:=0;i<out.Bounds().Max.X;i+=dx {
						
			var min float64 =999999
			var img *image.RGBA
			match:= -1
			var matches []int
			
			for v,mi := range tiles {	
									
				if mi.Uses<4 || coinFlip(){				
					var dif float64
					
					for dj:=0;dj<dy;dj++ {
						for di:=0;di<dx;di++ {							
							pixel := out.RGBAAt(i+di,j+dj)
							tpixel := mi.Tile.RGBAAt(di,dj)
							dif+= colorDistance(&tpixel, &pixel)							
						}
					}			
											
					if dif<min {
						match = v
						min = dif
					}
					
					if dif<maxColorDifference {
						matches=append(matches,v)
					}				
				}				
			}
			
			if len(matches)>0 {
				match = matches[rand.Intn(len(matches))]
			}
			
			img = tiles[match].Image
			tiles[match].Uses++
							
			draw.Draw(mosaic, image.Rect(tileScale*i/tileXResolution,tileScale*j/tileYResolution,tileScale*i/tileXResolution+tileScale,tileScale*j/tileXResolution+tileScale), img, img.Bounds().Min, draw.Src)
					
		}
	}	
	return mosaic
}

func buildMosaic(mr *mosRequest) *image.RGBA{
		
	var src *image.RGBA = mr.Image
	maxPixels := 480000
	
	for src.Bounds().Max.X*src.Bounds().Max.Y>maxPixels {
		src = downsample(src, image.Rect(0,0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2))
	}
		
	
	mr.Progress<-"downloading source images"
	urls:= flickrSearch(500,mr.Terms...)
	
	tiles:=downloadImages(urls)
	mr.Progress<-"building mosaic"
	
	
		
	return fitMosaic(src,tiles)
}

func main(){
	
	runtime.GOMAXPROCS(2)
		
  go func() {
				
	  for {
	    //var mr *MosRequest;
      select {
    	case mr := <-mosQueue:
			
				mr.Start = time.Now()
				mosaic:=buildMosaic(mr)
				mr.End= time.Now()
				
				fmt.Println("elapsed time: ", mr.End.Sub(mr.Start))
				mr.Progress<-"downloading mosaic"
				
				if mr.Save {
					fmt.Println("saving mosaic...")
					saveJPG(mosaic,"static/images/"+mr.Key+".jpg")
					savedMosaics=append(savedMosaics,mr.Key+".jpg")
					thumb:=downsample(mosaic,image.Rect(0,0,300,300))
					saveJPG(thumb,"static/thumbs/"+mr.Key+".jpg")
				}
				mr.Result<-mosaic
      }
	  }
 	}()
	
 
	http.HandleFunc("/postimage", postImage)
	http.HandleFunc("/listen", listen)
	http.HandleFunc("/api/images",getImages)
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.ListenAndServe(":555", nil)

}

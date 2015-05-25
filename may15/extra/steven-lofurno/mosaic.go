package main

import (
	"time"
	"math"
	"strconv"
	"io/ioutil"
	"net/http"
	"encoding/json"
		"image"
	_ "image/png"
	_ "image/jpeg"
		"log"
	"image/color"
)

//holds all information regarding the state of our mosaic request
type mosRequest struct {
	
	Image *image.RGBA
	Key string
	Id int
	Terms []string
	Progress chan string
	Result chan *image.RGBA
	Save bool
	Start time.Time
	End time.Time
}

type imageResponse struct{
	
		Image image.Image
		Err error	
}

type mosImage struct {
	
	Image *image.RGBA
	AvgColor *color.RGBA
	Tile *image.RGBA
	Uses int
}

type flickrPhoto struct{
	Id string `json:"id"`
	Owner string `json:"owner"`
	Secret string `json:"secret"`
	Server string `json:"server"`
	Farm int `json:"farm"`
	Title string `json:"title"`
	
}

func (p flickrPhoto) downloadURL() string {
		
	return "https://farm"+strconv.Itoa(p.Farm)+".staticflickr.com/"+p.Server+"/"+p.Id+"_"+p.Secret+".jpg"
	
}

type flickrResponsePhotos struct{
	
	Page int `json:"page"`
	Pages int `json:"pages"`
	Perpage int `json:"perpage"`
	Total string `json:"total"`
	Photo []flickrPhoto `json:"photo"`

	
}

type flickrResponse struct{
	Photos flickrResponsePhotos `json:"photos"`
	Stat string `json:"stat"`
	
}

func flickrSearch(count int, terms ...string) []string {
	
	results:= make([]string,0,count*len(terms))
	
	for _,term:=range terms {
		uri:="https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=749dec8d6d00d4df46215bf86e704bb0&text="+term+ "&page=1&format=json&per_page="+strconv.Itoa(count) + "&content_type=1&sort=relevance"
		res, err:= http.Get(uri)
		contents, err := ioutil.ReadAll(res.Body)
		rawlen := len(contents)
		
		j:=contents[14:rawlen-1]
			
		var f flickrResponse
		err=json.Unmarshal(j,&f)
		
		if err!=nil{
			log.Println(err)
		}
		
		for _,p:=range f.Photos.Photo {
			results = append(results,p.downloadURL())
			
		}
			
	}	
	return results
}

func downloadImages(urls []string) []mosImage {
	
	queue := make(chan string, 200)
  results := make(chan imageResponse, 200)
	
	for i := 0; i < 200; i++ {
    go imageDownloader(queue, results)
  }
	
	go func(){
		for _,url:=range urls{
			queue<-url
		}
		close(queue)
	}()
	
	var result imageResponse
	var images []mosImage
	
	for i := 0;i<len(urls);i++ {
		result= <-results
		
		if result.Err == nil {
			mi:=newMosImage(result.Image)
			images = append(images,mi)								
		}		
	}
	
	return images
	
}

func downloadAndDecode(url string) (image.Image, error){
	res, err := http.Get(url)
	defer res.Body.Close()
	
	m, _, err := image.Decode(res.Body)
	
	if err!=nil {
		log.Println(err)
	}
	
	return m,err
	
}

//accepts a source image and a sub-rectangle, and returns the average of
//the rgb colors within this rectangle
func averageColor(img *image.RGBA, rect image.Rectangle) color.RGBA {
	
	var rSum float64
	var gSum float64
	var bSum float64
	var count float64
	
	pixels := img.Pix
	
	stride:=img.Bounds().Max.X
	
	for i:=rect.Min.X; i<rect.Max.X;i++{
		for j:=rect.Min.Y;j<rect.Max.Y;j++{
			offset:=4*(j*stride+i)			
			
			rSum+= sRGBtoLinear(pixels[offset])
			gSum+= sRGBtoLinear(pixels[offset+1])
			bSum+= sRGBtoLinear(pixels[offset+2])

			count++			
		}
	}
			
	return color.RGBA{lineartosRGB(rSum/count), lineartosRGB(gSum/count), lineartosRGB(bSum/count), 255}
}

//downsamples a source image as close as possible to a desired size, while also
//maintaining a linear downsampling ratio. 
func downsample(img *image.RGBA, size image.Rectangle) *image.RGBA {
	
	xratio := int(img.Bounds().Max.X/size.Max.X)
	yratio:=int(img.Bounds().Max.Y/size.Max.Y)
	
	minratio:=xratio
	
	if yratio<xratio {
		minratio=yratio
	}
	
	xoffset:= int((img.Bounds().Max.X - size.Max.X*minratio)/2)
	yoffset:= int((img.Bounds().Max.Y - size.Max.Y*minratio)/2)
		
	out := image.NewRGBA(size)
	pixels := out.Pix
	
	for i:=0; i<size.Max.X;i++{
		for j:=0;j<size.Max.Y;j++{
			offset:=4*(j*size.Max.X+i)			
			r:=image.Rect(i*minratio+xoffset, j*minratio+yoffset, (i+1)*minratio+xoffset, (j+1)*minratio+yoffset)
					
			c:=averageColor(img, r)
			
			pixels[offset]=c.R
			pixels[offset+1]=c.G
			pixels[offset+2]=c.B
			pixels[offset+3]=255			
		}
	}	
	return out
}

/*
func lum(c *color.RGBA) float64 {
	return float64(0.299*float32(c.R) + 0.587*float32(c.G) + 0.114*float32(c.B))
}

func averageLum(img *image.RGBA, r image.Rectangle) float32 {
	
	c:=averageColor(img, r)	
	lum:=0.299*float32(c.R) + 0.587*float32(c.G) + 0.114*float32(c.B)
		
	return lum	
}
*/

//converts any type implmenting the image interface to a RGBA representation
func convertToRGBA(src image.Image) *image.RGBA {
	
	dst:=image.NewRGBA(src.Bounds())
	pix:=dst.Pix

	c:=0
	for j:=0;j<src.Bounds().Max.Y;j++ {		
		for i:=0;i<src.Bounds().Max.X;i++ {

			r1,g1,b1,_ := src.At(i,j).RGBA()
			
			pix[4*c]=uint8(r1)
			pix[4*c+1]=uint8(g1)
			pix[4*c+2]=uint8(b1)
			pix[4*c+3]=255
			
			c++			
		}		
	}	
	return dst
}

//casts or converts any type implmenting interface image to the RGBA type
func convertImage(m image.Image) *image.RGBA {
	
	var rgba *image.RGBA

	switch m.(type) {
	case *image.RGBA: 
		rgba=m.(*image.RGBA)
	default:
		rgba=convertToRGBA(m)
	}
	
	return rgba
}


func newMosImage(img image.Image) (mosImage) {
	//TODO assert image >= 64x64
	rgba := convertImage(img)
	
	var mi mosImage
	
	down:=downsample(rgba,image.Rect(0,0,64,64))
	tile:=downsample(down,image.Rect(0,0,2,2))
	
	mi=mosImage{}
	mi.Image=down
	mi.Tile=tile
	ac:=averageColor(down,down.Bounds())
	mi.AvgColor=&ac
	mi.Uses=0
	
	return mi	
}


func lineartosRGB(L float64) uint8 {
	
	var S float64
	var exp float64 = 1/2.4
	
	if L > 0.0031308 {
		S = 1.055*math.Pow(L,exp)-0.055
		
	} else {
		S = L * 12.92
	}
	
	return uint8(255*S)
}

func sRGBtoLinear(s uint8) float64 {
	
	var z float64 = float64(s)/255
		
	var L float64
	
	if z > 0.04045 {
		L = math.Pow((z + 0.055)/(1.055), 2.4)
	} else { 
		L = z/12.92
	}
	
	return L
}

func colorDistance3(e1 *color.RGBA, e2 *color.RGBA) float64 {
	
	r:=sRGBtoLinear(e1.R)-sRGBtoLinear(e2.R)
	g:=sRGBtoLinear(e1.G)-sRGBtoLinear(e2.G)
	b:=sRGBtoLinear(e1.B)-sRGBtoLinear(e2.B)
	
	return math.Sqrt(r*r+g*g+b*b)
	
}

func colorDistance(e1 *color.RGBA, e2 *color.RGBA) float64 {
	
	r:=float64(e1.R)-float64(e2.R)
	g:=float64(e1.G)-float64(e2.G)
	b:=float64(e1.B)-float64(e2.B)
	
	return math.Sqrt(r*r+g*g+b*b)
	
}

func colorDistance2(e1 *color.RGBA, e2 *color.RGBA) float64 {
	//http://www.compuphase.com/cmetric.htm
  rmean := int64(( e1.R + e2.R ) / 2)
  r := int64(e1.R - e2.R);
  g := int64(e1.G - e2.G);
  b := int64(e1.B - e2.B);
	
  return math.Sqrt(float64(  (((512+rmean)*r*r)>>8) + 4*g*g + (((767-rmean)*b*b)>>8)) );
}


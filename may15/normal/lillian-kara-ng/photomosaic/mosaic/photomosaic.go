package mosaic

import (
	"bitbucket.org/lillian_ng/photomosaic/descriptors"
	"bitbucket.org/lillian_ng/photomosaic/fetcher"
	"fmt"
)

// descriptor worker
func TileWorker(id int, builder descriptors.DescriptionBuilder, jobs <-chan *fetcher.ImageMetaData, results chan<- *Tile, done chan bool) {
	for j := range jobs {
		//fmt.Println("worker", id, "processing job", j.IgId)

		tile, err := GetTile(builder, j.SmallSizeUrl)
		if err != nil {
			done <- true
		} else {
			results <- &tile
		}
	}
}

func GetTile(b descriptors.DescriptionBuilder, url string) (Tile, error) {
	var t Tile
	t.Url = url

	m, err := fetcher.GetImage(url)
	if err != nil {
		return t, err
	}

	t.Description, err = b.GetDescription(m, m.Bounds())
	if err != nil {
		return t, err
	}

	return t, nil
}

// mosaic worker
func Matcher(m *Mosaic, tiles <-chan *Tile, done chan bool) {
	for t := range tiles {
		for i := range m.Descriptions {
			score, _ := m.Descriptions[i].MatchScore(t.Description)
			if m.MatchScores[i] > score {
				m.Urls[i] = t.Url
				m.MatchScores[i] = score
			}
		}
		done <- true
	}
}

func BuildMosaic(targetUrl, hashtag, clientId string) string {
	fmt.Println("Start mosaic...")
	sourceImageCount := 1000
	tileSize := 20
	builder := descriptors.NHistogramBuilder{2, 2}

	targetImage, _ := fetcher.GetImage(targetUrl)
	myMosaic := NewMosaic(targetImage.Bounds(), tileSize)
	myMosaic.FillDescriptions(targetImage, targetImage.Bounds(), builder)
	fmt.Println("My Mosaic will be", myMosaic.TilesWide, "x", myMosaic.TilesHigh)

	imgMeta := make(chan *fetcher.ImageMetaData, 100)
	tiles := make(chan *Tile, 100)
	done := make(chan bool)

	client := fetcher.NewInstagramClient(clientId)

	go client.Search(sourceImageCount, hashtag, imgMeta)

	for w := 0; w < 100; w++ {
		go TileWorker(w, builder, imgMeta, tiles, done)
	}

	go Matcher(&myMosaic, tiles, done)

	for i := 0; i < sourceImageCount; i++ {
		<-done
	}

	fmt.Println("DONE!")

	//file, _ := os.Create("mosaic.html")
	//fmt.Fprintf(file, MyMosaic.ToHTML())
	return myMosaic.ToHTML(tileSize)
}

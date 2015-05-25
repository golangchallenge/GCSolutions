package mosaic

import (
	"bitbucket.org/lillian_ng/photomosaic/descriptors"
	"bitbucket.org/lillian_ng/photomosaic/tile"
	"fmt"
	"image"
	"math"
)

type Tile struct {
	Url         string
	Description descriptors.Description
}

type Mosaic struct {
	Grid                 []image.Rectangle
	Urls                 []string
	Descriptions         []descriptors.Description
	MatchScores          []float64
	TilesWide, TilesHigh int
}

func NewMosaic(bounds image.Rectangle, tileSize int) Mosaic {
	var m Mosaic
	m.Grid, m.TilesWide = tile.ByPixel(bounds, tileSize, tileSize)
	m.TilesHigh = len(m.Grid) / m.TilesWide

	m.Urls = make([]string, len(m.Grid))
	m.MatchScores = make([]float64, len(m.Grid))
	m.Descriptions = make([]descriptors.Description, len(m.Grid))

	return m
}

func (m *Mosaic) FillDescriptions(targetImage image.Image,
	bounds image.Rectangle, builder descriptors.DescriptionBuilder) {
	for i := range m.Descriptions {
		description, err := builder.GetDescription(targetImage, m.Grid[i])
		if err != nil {
			fmt.Println("can't create description for tile")
		}
		m.Descriptions[i] = description
		m.MatchScores[i] = math.MaxFloat64
	}
}

func (m *Mosaic) ToHTML(tileSize int) string {
	var html string

	html = "<table cellpadding=0 cellspacing=0>\n"
	count := 0
	for i := range m.Urls {
		if count == 0 {
			html += "\t<tr>\n"
		}

		html += "\t\t<td><img src=\"" + m.Urls[i] + "\" "
		html += fmt.Sprintf("height='%v' width='%v'></td>\n", tileSize, tileSize)
		count++

		if count == m.TilesWide {
			html += "\t</tr>\n"
			count = 0
		}
	}
	return html
}

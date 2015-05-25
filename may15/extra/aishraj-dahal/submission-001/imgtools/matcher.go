package imgtools

import (
	"github.com/aishraj/gopherlisa/common"
	"image/color"
	"math"
)

func findClosestMatch(context *common.AppContext, inpVal color.RGBA, group string) string {
	context.Log.Println("Starting to find closest match. Params are :", inpVal, group)
	db := context.Db
	stmt, err := db.Prepare("SELECT img,red,green,blue FROM Images WHERE imgtype = ? LIMIT 10")
	if err != nil {
		context.Log.Fatal("ERROR: Cannot get images from db. ", err)
		return ""
	}
	defer stmt.Close()
	minDistance := math.MaxFloat64
	minFileName := ""
	results, err := stmt.Query(group)
	if err != nil {
		context.Log.Fatal("Could not execute statment on db")
		return ""
	}
	var imageName string
	var redVal uint8
	var greenVal uint8
	var blueVal uint8

	for results.Next() {
		err := results.Scan(&imageName, &redVal, &greenVal, &blueVal)
		context.Log.Printf("The values are image name %v, red val %v, green val %v, blue val %v", imageName, redVal, greenVal, blueVal)
		if err != nil {
			context.Log.Fatal("Unable to get the canend result into the table.", err)
			return ""
		}
		//Values are based on a stackoverflow answer. They seem to work somehow.
		//Here http://j.mp/1IOpkRV
		rFactor := 0.30
		gFactor := 0.59
		bFactor := 0.11

		rDiff := float64(inpVal.R - redVal)
		gDif := float64(inpVal.G - greenVal)
		bDiff := float64(inpVal.B - blueVal)

		diff := math.Pow((rDiff*rFactor), 2) + math.Pow((gDif*gFactor), 2) + math.Pow((bDiff*bFactor), 2)
		context.Log.Println("Diff is", diff)

		if diff < minDistance {
			minDistance = diff
			minFileName = imageName
			context.Log.Println("Min is now", minDistance)
		}
	}
	context.Log.Println("The closest fileName is: ", minFileName)
	return minFileName
}

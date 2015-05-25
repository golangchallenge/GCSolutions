# Go-Mosaic

Go-Mosaic is a mosaic creator written in Go.

This was written as an entry for [Go Challenge 3](http://golang-challenge.com/go-challenge3/).

## Installation

    go get github.com/valentine/go-mosaic/

## Usage

    go-mosaic -input="input.jpg" -output="output.jpg" source="photodirectory" -tile=16    
    
**input**: Path of input image (used to create mosaic)

**output**: Path of output image (defaults to **output.jpg**)

**source**: Directory for source images (defaults to **photos/**)

**tile**: Size of tiles in output image (higher numbers process faster, but images will have less resolution) (defaults to **16**)
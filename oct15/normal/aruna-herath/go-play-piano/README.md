## Go Play piano

### To run with source or build..

The only dependency apart from the Go standard packages is [mgl32](https://github.com/go-gl/mathgl)

`go get -u github.com/go-gl/mathgl`

This is needed to calculate view and projection matrices. I am thinking about hardcoding those matrices and remove the dependancy.

`go run main.go key.go board.go sound.go`

`gomobile build` 

all: clean deps build

build:
	go fmt
	go build
test:
	go test
deps:
	go get -d -t
import: tools
	goimports -w .
tools:
	go get golang.org/x/tools/cmd/goimports
clean:
	rm -f ./sudoku ./sudoku.zip
bench:
	go test -bench=.
cover:
	go test -cover
coverhtml:
	go test --coverprofile=cover.out
	go tool cover -html=cover.out
	rm -f cover.out
coverfunc:
	go test --coverprofile=cover.out
	go tool cover -func=cover.out
	rm -f cover.out
install: clean deps build
	go install
package: clean deps build
	zip sudoku.zip *.go README.md LICENSE sample Makefile sudoku

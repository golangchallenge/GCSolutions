all: check test install

check: goimports govet

goimports:
	@goimports -d .

govet:
	@go tool vet -all .

test:
	@go test -v -cover

bench:
	@go test -v -run=dummy -bench=.

install:
	@go install

clean:
	@-rm -v "sudoku" 2>/dev/null
	@-rm -v "$(GOPATH)/bin/sudoku" 2>/dev/null

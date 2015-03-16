BUILD_ARGS=
DEBUG_ARGS= -gcflags "-N -l"

# List of directories is needed in order to avoid running checks against vendor
# If your project has multiple packages, consider moving them into a version
# directory. For example v1/config, v1/handlers, ...
DIRS := .

default: build

debug: BUILD_ARGS+=$(DEBUG_ARGS)
debug: build

# TODO Enable golint here
build: deps fmt get errcheck vet lint test
	go build $(BUILD_ARGS)

fmt: 
	go fmt $(DIRS:%=github.com/kellydunn/go-challenge-1/%/...)

get:
	go get

test:
	go test github.com/kellydunn/go-challenge-1

bench:
	go test -run=XXX -bench=. github.com/kellydunn/go-challenge-1

errcheck: deps
	errcheck $(DIRS:%=github.com/kellydunn/go-challenge-1/%/...)

lint: deps
	golint $(DIRS) 

vet: deps
	go vet $(DIRS:%=github.com/kellydunn/go-challenge-1/%/...)

deps:
	go get github.com/kisielk/errcheck
	go get github.com/golang/lint/golint

clean:
	go clean

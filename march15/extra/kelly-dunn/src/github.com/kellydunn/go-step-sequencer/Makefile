BUILD_ARGS=
DEBUG_ARGS= -gcflags "-N -l"

# List of directories is needed in order to avoid running checks against vendor
# If your project has multiple packages, consider moving them into a version
# directory. For example v1/config, v1/handlers, ...
DIRS := sequencer

default: build

debug: BUILD_ARGS+=$(DEBUG_ARGS)
debug: build

# TODO Enable golint here
build: deps fmt get vet lint test
	go build $(BUILD_ARGS)

fmt: 
	go fmt $(DIRS:%=github.com/kellydunn/go-step-sequencer/%/...)

get:
	go get

test:
	go test github.com/kellydunn/go-step-sequencer
	go test $(DIRS:%=github.com/kellydunn/go-step-sequencer/%/...)

bench:
	go test -run=XXX -bench=. github.com/kellydunn/go-step-sequencer

errcheck: deps
	errcheck $(DIRS:%=github.com/kellydunn/go-step-sequencer/%/...)

lint: deps
	golint $(DIRS) 

vet: deps
	go vet $(DIRS:%=github.com/kellydunn/go-step-sequencer/%/...)

deps:
	go get github.com/kisielk/errcheck
	go get github.com/golang/lint/golint

clean:
	go clean

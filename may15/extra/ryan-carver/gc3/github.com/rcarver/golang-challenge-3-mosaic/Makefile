# Build

build:
	go install github.com/rcarver/golang-challenge-3-mosaic/mosaicly

.PHONY: build

# Use

gen: build
	rm -f output.jpg
	$$GOPATH/bin/mosaicly fetch -tag balloon
	$$GOPATH/bin/mosaicly gen -tag balloon -in fixtures/balloon.jpg -out output.jpg
	test -f output.jpg && open output.jpg

serve: build
	$$GOPATH/bin/mosaicly serve

.PHONY: gen serve

# Sample

sample.jpg: build
	$$GOPATH/bin/mosaicly fetch -tag balloon -num 2000
	$$GOPATH/bin/mosaicly gen -tag balloon -in fixtures/balloon-square.jpg -out $@
	test -f $@ && open $@

# Lint

lint:
	go fmt ./...
	go vet ./...
	$$GOPATH/bin/golint ./...

.PHONY: lint

# Test

test: test_unit test_cli test_service 

test_unit:
	go test ./...

test_cli: build
	./tests/cli.sh

test_service: build
	./tests/service.sh

.PHONY: test test_unit test_cli test_service

# Test Coverage

cov_packages=mosaic instagram
cov_files=$(addsuffix .cov,$(cov_packages))
cov_html=$(addsuffix .cov.html,$(cov_packages))

cov: clean_cov $(cov_files) $(cov_html)

clean_cov: 
	rm -f $(cov_files)

%.cov:
	go test -coverprofile=$@ ./$(firstword $(subst ., ,$@))

%.cov.html:
	go tool cover -html=$(subst .html,,$@)

.PHONY: cov clean_cov


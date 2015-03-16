#!/bin/bash

go vet . && golint decoder.go drum.go && go test -v .
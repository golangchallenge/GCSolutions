#!/bin/sh

go build
./gc4 -generate=3 | ./gc4

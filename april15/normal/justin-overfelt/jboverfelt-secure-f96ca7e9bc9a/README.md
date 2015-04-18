# Go Challenge 2

Securing data transmission using NaCl.

More information [here](http://golang-challenge.com/go-challenge2/)

## Prerequisites

Install Go's NaCl library:

```
go get -u golang.org/x/crypto/nacl/box
```

## Installation/Usage

Please place the files in this directory in ``$GOPATH/src/bitbucket.org/jboverfelt/secure``

``go get`` will not work because this is hosted in a private repository until the challenge is complete

To build the included command, change to the cmd/challenge2 directory and run ``go build``

Tests were split up into two files, one for the library and one for the command

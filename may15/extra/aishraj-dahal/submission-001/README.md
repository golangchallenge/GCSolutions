## Gopherlisa

gopherlisa is a mosaic generator built as a part of [Go challenge](http://golang-challenge.com/). Even though, the objective of the challenge changed from building a web application to building a standalone mosaic generator, as I was already halfway developing the web application, I decided to continue on it and built it into a web application that generates mosaic by using images from Instagram.

### Getting started

In order to get started with setting up the package, you'll need to have the following installed:
1. Go development environment
2. MySQL

Based on your platform  (Linux/OS X), once you've setup the above, you'll need to get the Instagram Client ID and Instagram Secret before launching the app. In order to get the necessary credentials please read this [link](https://instagram.com/developer/clients/manage/).
Note that while retrieving the Instagram Client ID, you might be prompted for a redirect URI. This URI has be publically accessible to authenticate into Instagram. In case you're using it a local machine, I suggest setting up a local tunnel using [ngrok](https://ngrok.io). Once you're done with this, please update the `gopherlisa_constants.go` file with the values you just got.

Also, enter the relevant MySQL credentials on the same file.

The location where the files will be downloaded from Instagram can be configured changing the constant `DownloadBasePath` in `gopherlisa_constants.go`. Note that the directory must exist.

(*while I think its not a good practice to use a file with consts to read configs , the Go challenge 3 did not allow any other 3rd party libraries to be used for config management. JSON could have been a good way to go, but then again, I prefer TOML over it.*)

#### Running/Building

`go get github.com/aishraj/gopherlisa`
`go build github.com/aishraj/gopherlisa`

In order to run the program, launch the binary `./gopherlisa`. This would start a web server which would take you to a login page, allowing you to login using Instagram.

####Known issues

The quality of the generated mosaic can vary from average to ridiculously bad. I've had no prior experience with Image processing and this being my only second Go project, may explain a lot of it.

Also to add, the database lookup for the closest image is quite slow. The search may return the same image multiple times and the resulting mosaic might be possibly composed of only a handful of images.

The session management does not cleanup timed out sessions.

#### Disclaimer
This project is a bad example for writing good and performance critical Go code. If you're looking at this project as a reference, I suggest you to checkout [Effective Go](https://golang.org/doc/effective_go.html) instead.

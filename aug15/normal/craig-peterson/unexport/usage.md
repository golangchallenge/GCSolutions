## Unexport - A tool for finding and getting rid of unneccesary exports.

Usage:

Most basic usage is `unexport pkg`, where package should be the fully qualified package name (`github.com/myname/whatever`). Unexport will find that package in your gopath, find all exported symbols, analyze their usages on your gopath, and output `gorename` commands to stdout to unexport uneeded symbols. You can save those commands, filter them to your needs, and run them.

If you want to simply run all rename operations in one shot, you can force the run with `unexport -f pkg`.

If you want verbose logging, you can use `-v`.

Sometimes it is difficult to know what packages are using yours. Especially if it is publicly availible. The `-d` flag will use godoc.org's api to find all known importers of your package. It will attempt to download them with `go get -d -v`, and then exit. This will make sure as many usages as possible are availible on your gopath.
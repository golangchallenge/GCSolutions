# unexport

Unexport is a tool which unexports nonused exported identifiers. It was created
as a part of Go challenge [#5](http://golang-challenge.com/go-challenge5/).
Checkout the examples below


## Install

```bash
go get github.com/fatih/unexport
```

## Usage and Examples

Unexport exported identifiers in the encoding/pem package:

```
$ unexport -package encoding/pem
```

Process the package and display any possible changes. It doesn't unexport
exported identifiers because of the -dryrun flag
  
```
$ unexport -package github.com/fatih/color -dryrun
```
  
Unexport only the "Color" and "Attribute" identifiers from the
github.com/fatih/color package. Note that if the identifiers are used by
other packages, it'll silently fail
  
```
$ unexport -package github.com/fatih/color -identifier "Color,Attribute"
```
  
## License

The BSD 3-Clause License - see LICENSE for more details


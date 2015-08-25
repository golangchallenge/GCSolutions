Unexport
--------
**This is an entry to the The August 2015 Go Challenge a.k.a. The Go Challenge 5**

Unexport scans a package for exported identifiers, it then parses other packages on the `$GOROOT` and `$GOPATH` to see if any other packages have incoming references. If there are references which are not used by any other packages, unexport will rename these identifiers to an unexported equivalent through the `gorename` tool.

**Usage**

    $ ./unexport -pkg github.com/foo/bar

**Flags**

  -pkg string
    	Which package to scan for exported identifiers
  -safe
    	Make unexport only work on internal packages
  -unexportAll
    	Don't interactively ask for each unused exported identifier, just export all.

**Issues**
Seems to be an issue with go/types or with how unexport is using go/types, errors like the following are returned by go/types:

    cannot pass argument x (variable of type *cmd/compile/internal/big.Int) to parameter of type *cmd/compile/internal/big.Int

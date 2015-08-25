The import path should be github.com/robhor/gochallenge5.
I went a bit overboard with including all the rename-safety checking in this package, but it is faster than calling gorename for each unexported identifier, since it has to reload the build context each time.
It would be cool if the refactor/rename package exported a bit more reusable functionality, that would've made this a lot easier :)

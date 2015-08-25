# Run in Docker #

```

docker build -t go-unexport .
docker run -it --rm --name go-unexport go-unexport
docker run -it --rm -v /home/dooman/Projects/go/:/go -p 6060:6060  --name gounexport golang

```

# Thoughts #

It was a Friday and I was actually thinking what to do on the weekend, because
my wife was going with her friends to do some shopping. I was choosing between bike
ride on ocean road or have a garage party with friends. Why did I open reddit?
You guys did that weekend and next two :)

Well, I came with C/C++ and Java experience. My last 8 years was exclusively hold
by JVM and Web(yeah, still need to be full stack and agile). This application is my first
serious golang experience. There are some points from me:

* Easy to start. I supposed that it will be more like C. All this nightmare with
compilers and shared libraries. But it was really comfy, like Java. Also, I developed
using Docker and I think this way is the easiest. It's probably worth to *add link
with docker images* to "Getting Started" guide on golang.org.

* Tools are awesome. I decided to try new editors. I tried Atom and Sublime. They
are both good. I suppose, that the root of happiness is govet, gofmt, golint,...
I really like that you choose Unix way to develop language environment. It's
really brilliant idea.

* Docs are good. I didn't have any problems with starting develop gounexport tool.
Examples are super handy and is awesome idea to store them in code make them compiled!

* go/types package. It's pretty staightforward and easy to use. I had a problem
with importing builtin packages. I'm still thinking that I'm doing something
wrong, because I don't understand why method definitions have names like
(interface).Less instead of (sort.Interface).Less for builtin(compiled)
types.

* Errors handling. To be honest, this one is not the best experience. I remember
that it was a pain point in C, as well. Golang brought me back to that time.
I know that try-catch is not really nice from performance point and not clear
semantically, but I think it's more straightforward then process each error
separately.
Just, to illustrate, fs.ReplaceStringInFile:

```go
sourceFile, err := os.OpenFile(file, os.O_RDWR, 0)
if err != nil {
  return err
}

var info os.FileInfo
var restFile []byte
seekTo := offset + len(from)

if info, err = sourceFile.Stat(); err != nil {
  goto closeAndReturn
}

restFile = make([]byte, int(info.Size())-seekTo)
if _, err = sourceFile.Seek(int64(seekTo), 0); err != nil {
  goto closeAndReturn
}
if _, err = sourceFile.Read(restFile); err != nil {
  goto closeAndReturn
}
//...5 more calls to truncate file and replace string

if err = sourceFile.Close(); err != nil {
  goto closeAndReturn
}

closeAndReturn:
closeErr := sourceFile.Close()
if closeErr != nil && err == nil {
  err = closeErr
}

return err
```

* Tests. Good and easy. Included coverage is a nice bonus. Little bit miss
mock framework. Maybe for next gochallange's it would be nice to *allow
github.com/golang/mock/gomock* or other. Again, examples are excellent idea!

* Debugging. I didn't use debugger because it was *not in docker image (nice to have)*
and I didn't want to spend time to setup it. Anyway it's not really neccessary
for such application (one thread and not big enough).

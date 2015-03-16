## Submission to go-challenge #1
See [FAQ](http://golang-challenge.com/faq/) for more info about the go-challenge.

The goal of this challenge is to write a binary decoder that given a binary backup,
outputs the same printouts as in the [challenge](./challenge.md) defined. See
http://golang-challenge.com/go-challenge1/ for full description.

## Getting started
This example does not provide or generate any executable.
To run the tests:
~~~bash
go test -v
~~~

Should return
~~~bash
=== RUN TestDecodeFile
--- PASS: TestDecodeFile (0.00s)
PASS
ok  	...
~~~

### Assumptions and design decisions
* File Format
<pre>
|SPLICE (6 bytes)|Payload size (8 bytes)|                              => File Header
|Version (32 byte)| Tempo (4 bytes)|                                   => Pattern
|Id (4 bytes)|Name length (1 byte)| Name (n bytes)| Steps (16 bytes)|  => First Track
|Id (4 bytes)|Name length (1 byte)| Name (n bytes)| Steps (16 bytes)|  => Next Track
...
</pre>

* Files might be large. Therefore payload size uses 8 bytes (big endian) instead of empty bytes
and smaller type in the header.
* For simplicity `Pattern.String()` is used for the printout although this would be too verbose
when used with logging.
* The `Pattern.String()` implementation may not be as readable as a `text/template` but
performed better in benchmarks.

## Author
* Twitter: [@alpe1] (https://twitter.com/alpe1)

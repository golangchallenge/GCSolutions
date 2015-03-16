Go Challenge 1
==============

For more information about the challenge see http://golang-challenge.com/go-challenge1/.

## Test

```
$ go test -v
=== RUN TestDecodeFile
--- PASS: TestDecodeFile (0.00 seconds)
=== RUN TestEncodeFile
--- PASS: TestEncodeFile (0.00 seconds)
PASS
ok  	github.com/gedex/drum0.007s
```

## Benchmark

```
$ go test -bench=. -benchmem
PASS
BenchmarkDecodePattern1  500000      7383 ns/op    1460 B/op      38 allocs/op
BenchmarkDecodePattern2  500000      6490 ns/op    1016 B/op      27 allocs/op
BenchmarkDecodePattern3  500000      7397 ns/op    1447 B/op      37 allocs/op
BenchmarkDecodePattern4  500000      6914 ns/op    1057 B/op      29 allocs/op
BenchmarkDecodePattern5  500000      5137 ns/op     636 B/op      18 allocs/op
BenchmarkEncodePattern1  200000     14514 ns/op    3526 B/op      82 allocs/op
BenchmarkEncodePattern2  200000     10929 ns/op    2676 B/op      60 allocs/op
BenchmarkEncodePattern3  200000     14421 ns/op    3512 B/op      82 allocs/op
BenchmarkEncodePattern4  200000     10734 ns/op    2690 B/op      61 allocs/op
BenchmarkEncodePattern5  500000      6767 ns/op    1487 B/op      37 allocs/op
ok  	github.com/gedex/drum070.653s

```

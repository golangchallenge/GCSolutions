# General
Go Challenge I Submission

# Author Details
    Name: Matt T. Proud
    Country of Residence: Zürich, Schweiz
    Twitter: khanreaper https://twitter.com/khanreaper
    Google Plus: https://plus.google.com/+MattProud
    URL: http://www.matttproud.com

# Installation

    go get github.com/matttproud/golang_challenge/I/drum

# API Examination

    godoc -http=":6060"

or

    godoc github.com/matttproud/golang_challenge/I/drum | ${PAGER}

# Tests

    go test -test.v=true github.com/matttproud/golang_challenge/I/drum

# Benchmarks

    go test -test.v=true -test.bench="Benchmark*" github.com/matttproud/golang_challenge/I/drum

Benchmarks may report what follows:

    BenchmarkDecodePattern1 100000  15257 ns/op 13.83 MB/s  2184 B/op 92 allocs/op
    BenchmarkDecodePattern2 200000  11620 ns/op 13.51 MB/s  1504 B/op 69 allocs/op
    BenchmarkDecodePattern3 100000  17248 ns/op 12.23 MB/s  2192 B/op 92 allocs/op
    BenchmarkDecodePattern4 100000  12228 ns/op 13.17 MB/s  1520 B/op 69 allocs/op
    BenchmarkDecodePattern5 200000  8456 ns/op  15.61 MB/s  976 B/op  46 allocs/op

The use of `binary.Read` and other facilities make implementation simple;
however, they are somewhat inefficient due to type switching and other
preemptory allocation.  Were we to care about this, we would manually
scan the `io.Reader` ourselves without using its automatic struct
population mechanism.

Another consideration were this performance-critical would be to create a
dedicated Decoder type that would contain its own buffers to amortize
allocation costs across operations.

# License
Apache License, Version 2.0

# Disclaimers and Miscellany
I, the entrant, shall indemnify, defend, and hold JoshSoftware Pvt. Ltd.
(who has sponsored the domain and is the organizer of these challenges)
harmless from any third party claims arising from or related to that
entrant's participation in the Challenge. In no event shall JoshSoftware
Pvt. Ltd. be liable to an entrant for acts or omissions arising out of or
related to the Challenge or that entrant's participation in the Challenge.

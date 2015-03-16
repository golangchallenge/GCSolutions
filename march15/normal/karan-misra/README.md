# 1 : Drum Decode

This was quite an interesting challenge. I always love optimizing stuff, and considering the problem statement and the weightage on a performant solution, I had a lot of fun with this.

The initial implementation looked like this:

	BenchmarkDecode	 200000      7400 ns/op       2642 B/op       42 allocs/op

The latest version looks like this:

	BenchmarkDecode	 1000000     2080 ns/op       592 B/op        3 allocs/op

(the numbers are from a MBP Late 2013, 2.3 Ghz Quad with 16 GB RAM)

This improvement has been achieved without twisting the original design. Just smart reuse of buffers and a few other tricks to work around Go's relatively immature escape analysis algorithms.

Having said that, I am sure that there is a lot of scope of improvement and possibly? even a idiomatic zero alloc version of the same (I did manage to get it down to 0 alloc/op but that implementation required me to do things which felt downright dirty).

## Instructions

To setup:

	./setup.sh

To test:

	./test.sh

To bench:

	./bench.sh

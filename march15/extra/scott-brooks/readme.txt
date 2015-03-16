This submissions also includes a splice to wav writer.  It can be found in cmd/splice2wav.go, and looks for the sample files inside the samples directory.  It should be run with the working directory the same as this file.

You can use it as follows

go run cmd/splice2wav.go fixtures/pattern_5.splice pattern5.wav

By default it will create a 20 second sample


For a longer sample, you can use the following for a 40 second sample
go run cmd/splice2wav.go -length=40 fixtures/pattern_5.splice pattern5.wav



The included wav files in the samples directory were downloaded from
http://smd-records.com/tr808/?page_id=14

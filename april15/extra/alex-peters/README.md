# Go Challenge #2
Submission to [Go challenge #2](http://golang-challenge.com/go-challenge2/).

## Goal of the challenge

   In order to prevent our competitor from spying on our network, we are going to write a small system that leverages NaCl to establish secure communication. NaCl is a crypto system that uses a public key for encryption and a private key for decryption.
   Your goal is to implement the functions in main.go and make it pass the provided tests in main_test.go.


In order to test our echo server, we can do:

~~~bash
$ go build
$ ./challenge-02 -l 8080&
$ ./challenge-02 8080 "hello world"
hello world
~~~

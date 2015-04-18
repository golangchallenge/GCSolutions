package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"gochallenge2/secureconn"
)

const usage = `usage:
Create a secure echo server:
	gochallenge2 -l [host:]<port>

Send a message to secure server:
	gochallenge2 [host:]<port> <message>
`

var listen = flag.Bool("l", false, "Listen mode")

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return secureconn.NewSecureReader(r, priv, pub)
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return secureconn.NewSecureWriter(w, priv, pub)
}

// Dial connects to an address using tcp and returns a secured connection.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return secureconn.New(conn)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			defer c.Close()

			sc, err := secureconn.New(c)
			if err != nil {
				return
			}
			defer sc.Close()

			for {
				_, err := io.Copy(sc, sc)
				if err == io.EOF {
					return
				}
				if err != nil {
					log.Printf("secure echo server: error serving %v: %v\n",
						c.RemoteAddr(), err)
					return
				}
			}
		}(conn)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		os.Exit(2)
	}

	flag.Parse()
	if (flag.NFlag() == 0 && flag.NArg() != 2) || (flag.NFlag() == 1 && flag.NArg() > 1) {
		flag.Usage()
	}

	addr := flag.Arg(0)
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	// Server mode
	if *listen {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()

		log.Fatal(Serve(l))
	}

	// Client mode
	conn, err := Dial(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := []byte(flag.Arg(1))
	if _, err := conn.Write(buf); err != nil {
		log.Fatal(err)
	}

	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", buf[:n])
}

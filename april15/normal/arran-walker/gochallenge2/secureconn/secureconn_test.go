package secureconn

import (
	"bytes"
	"crypto/rand"
	"io"
	"net"
	"testing"
)

func TestSteamingLargeData(t *testing.T) {
	r, w := io.Pipe()

	conn, err := New(struct {
		io.Reader
		io.Writer
		io.Closer
	}{r, w, w})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send increasingly large data packets:
	// 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144 (bytes)
	for i := uint(1); i <= 10; i++ {
		outgoing := make([]byte, 256<<i)
		incoming := make([]byte, len(outgoing))

		io.ReadFull(rand.Reader, outgoing)
		go func() {
			if _, err := conn.Write(outgoing); err != nil {
				t.Fatal(err)
			}
		}()

		_, err := io.ReadFull(conn, incoming)
		if err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(outgoing, incoming) != 0 {
			t.Fatalf("Large data of size %d did not match", 256<<i)
		}
	}
}

func TestStreamingChunkSize(t *testing.T) {
	r, w := io.Pipe()

	conn, err := New(struct {
		io.Reader
		io.Writer
		io.Closer
	}{r, w, w})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Set chunk size too small
	err = conn.SetStreamingChunkSize(1)
	if err == nil {
		t.Fatal("Streaming chunk size was incorrectly set too small")
	}

	// Set chunk size too large
	err = conn.SetStreamingChunkSize(666666)
	if err == nil {
		t.Fatal("Streaming chunk size was incorrectly set too large")
	}

	// Set chunk size just right
	err = conn.SetStreamingChunkSize(2048 - Overhead)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReplayPrevention(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server
	go func(l net.Listener) {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()

				sc, err := New(c)
				if err != nil {
					t.Fatal(err)
				}
				defer sc.Close()

				buf := make([]byte, 2048)
				for {
					_, err := sc.Read(buf)

					// Check we correctly receive a replay attack error
					if err != nil && err != ErrReplayAttack {
						t.Fatalf("Unexpected error (%s). Expected ErrReplayAttack", err)
					}
					if err == ErrReplayAttack {
						return
					}
				}
			}(conn)
		}
	}(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Create buffer for intercepted messages
	intercept := bytes.NewBuffer(nil)

	// Create a proxied connection that writes data to the intercepter
	proxy, err := New(&struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		conn,
		intercept,
		conn,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Write the first intercepted packet (the key) to the underlying transport
	_, err = intercept.WriteTo(conn)
	if err != nil {
		t.Fatal(err)
	}

	// Send replay message to be intercepted
	_, err = proxy.Write([]byte("replay"))
	if err != nil {
		t.Fatal(err)
	}

	// Write intercepted message twice to underlying transport
	_, err = conn.Write(append(intercept.Bytes(), intercept.Bytes()...))
	if err != nil {
		t.Fatal(err)
	}

	io.Copy(nil, proxy)
}

func TestTampering(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server
	go func(l net.Listener) {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()

				sc, err := New(c)
				if err != nil {
					t.Fatal(err)
				}
				defer sc.Close()

				buf := make([]byte, 2048)
				for {
					_, err := sc.Read(buf)

					// Check we correctly receive a verification error
					if err != nil && err != ErrVerification {
						t.Fatalf("Unexpected error (%s). Expected ErrVerification", err)
					}
					if err == ErrVerification {
						return
					}
				}
			}(conn)
		}
	}(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Create buffer for intercepted messages
	intercept := bytes.NewBuffer(nil)

	// Create a proxied connection that writes data to the intercepter
	proxy, err := New(&struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		conn,
		intercept,
		conn,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Write the first intercepted packet (the key) to the underlying transport
	_, err = intercept.WriteTo(conn)
	if err != nil {
		t.Fatal(err)
	}

	// Send replay message to be intercepted
	_, err = proxy.Write([]byte("tampered"))
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with data
	// Replace packet count of 1 to 5
	intercept.Bytes()[6] = 0x05

	// Send tampered data
	_, err = conn.Write(intercept.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	io.Copy(nil, proxy)
}

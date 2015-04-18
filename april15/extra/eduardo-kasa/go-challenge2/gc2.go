package gc2

import (
	"io"
	"net"
	"time"
)

type timeoutError struct{}

func (timeoutError) Error() string   { return "bla: DialWithDialer timed out" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	return DialWithDialer(new(net.Dialer), addr)
}

func DialWithDialer(dialer *net.Dialer, addr string) (io.ReadWriteCloser, error) {
	// We want the Timeout and Deadline values from dialer to cover the
	// whole process: TCP connection and the handshake. This means that we
	// also need to start our own timers now.
	timeout := dialer.Timeout

	if !dialer.Deadline.IsZero() {
		deadlineTimeout := dialer.Deadline.Sub(time.Now())
		if timeout == 0 || deadlineTimeout < timeout {
			timeout = deadlineTimeout
		}
	}

	var errChannel chan error

	if timeout != 0 {
		errChannel = make(chan error, 2)
		time.AfterFunc(timeout, func() {
			errChannel <- timeoutError{}
		})
	}

	rawConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn := Client(rawConn)

	if timeout == 0 {
		err = conn.Handshake()
	} else {
		go func() {
			errChannel <- conn.Handshake()
		}()

		err = <-errChannel
	}

	if err != nil {
		rawConn.Close()
		return nil, err
	}
	return conn, nil
}

func Client(conn net.Conn) *Conn {
	return &Conn{conn: conn, isClient: true}
}

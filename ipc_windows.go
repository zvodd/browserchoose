//go:build windows

package browserchoose

import (
	"errors"
	"io"
	"net"
	"strings"

	"github.com/Microsoft/go-winio"
)

// pipeName is how one-shot handler invocations reach the running daemon.
const pipeName = `\\.\pipe\browserchoose.zvodd.github.com`

const maxURLLen = 16 * 1024

// ForwardURL hands url to a running daemon. It returns false when no daemon is
// listening, in which case the caller should handle the URL itself.
func ForwardURL(url string) bool {
	conn, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		return false
	}
	defer conn.Close()
	_, err = io.WriteString(conn, url)
	return err == nil
}

// ListenURLs claims the pipe and passes each forwarded URL to handle on a
// background goroutine. It returns an error if another daemon is already
// listening.
func ListenURLs(handle func(url string)) error {
	if conn, err := winio.DialPipe(pipeName, nil); err == nil {
		conn.Close()
		return errors.New("another daemon is already listening")
	}
	listener, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go readURL(conn, handle)
		}
	}()
	return nil
}

func readURL(conn net.Conn, handle func(url string)) {
	defer conn.Close()
	raw, err := io.ReadAll(io.LimitReader(conn, maxURLLen))
	if err != nil {
		return
	}
	if url := strings.TrimSpace(string(raw)); url != "" {
		handle(url)
	}
}

package utils

import (
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// https://github.com/sausagenoods/snitch/blob/master/conn.go
// Mock io.Reader to satisfy the net.Conn interface
// https://www.agwa.name/blog/post/writing_an_sni_proxy_in_go
type MockConn struct {
	Reader io.Reader
}

func (conn MockConn) Read(p []byte) (int, error)         { return conn.Reader.Read(p) }
func (conn MockConn) Write(p []byte) (int, error)        { return 0, nil }
func (conn MockConn) Close() error                       { return nil }
func (conn MockConn) LocalAddr() net.Addr                { return nil }
func (conn MockConn) RemoteAddr() net.Addr               { return nil }
func (conn MockConn) SetDeadline(t time.Time) error      { return nil }
func (conn MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (conn MockConn) SetWriteDeadline(t time.Time) error { return nil }

//------------

type WrapperConn struct {
	net.Conn
	Reader io.Reader
}

func (conn WrapperConn) Read(p []byte) (int, error) { return conn.Reader.Read(p) }

func GetHost(buff []byte) (string, bool) {

	head_split := strings.SplitSeq(string(buff), "\r\n")

	for h := range head_split {
		host, ok := strings.CutPrefix(h, "Host: ")
		if ok {
			return host, ok
		}
	}

	return "", false
}

// copied from grok ai
// reality server require it
func (conn WrapperConn) CloseWrite() error {
	if cw, ok := conn.Conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return nil
}

///////

func Copy(c1, c2 net.Conn) {

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer c1.Close()
		defer wg.Done()

		io.Copy(c1, c2)
	}()

	go func() {
		defer c2.Close()
		defer wg.Done()

		io.Copy(c2, c1)
	}()

	wg.Wait()

}

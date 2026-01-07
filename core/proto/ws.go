package proto

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
	"tun/core"

	"github.com/gorilla/websocket"
)

// mostly copied from https://github.com/XTLS/Xray-core/blob/main/transport/internet/websocket/connection.go
type WsConn struct {
	conn *websocket.Conn
	r    io.Reader
}

func (c *WsConn) Read(b []byte) (int, error) {
	for {
		reader, err := c.getReader()
		if err != nil {
			return 0, err
		}

		nBytes, err := reader.Read(b)
		if err != nil {
			c.r = nil
			continue
		}
		return nBytes, err
	}
}

func (c *WsConn) getReader() (io.Reader, error) {
	if c.r != nil {
		return c.r, nil
	}

	_, reader, err := c.conn.NextReader()
	if err != nil {
		return nil, err
	}
	c.r = reader
	return reader, nil
}

func (c *WsConn) Write(b []byte) (int, error) {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

// close by websocket close message is cringe so i only close underlying conn
func (c *WsConn) Close() error {
	return c.conn.UnderlyingConn().Close()
	// var errs []any
	//
	//	if err := c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second*5)); err != nil {
	//		errs = append(errs, err)
	//	}
	//
	//	if err := c.conn.Close(); err != nil {
	//		errs = append(errs, err)
	//	}
	//
	//	if len(errs) > 0 {
	//		return errors.New("failed to close connection")
	//	}
	//
	// return nil
}

// underlying conn local addr
// the actual impl in xray source throws error
func (c *WsConn) LocalAddr() net.Addr {
	return c.conn.UnderlyingConn().LocalAddr()
}

// underlying conn remote addr
// the impl in xray source throws error
func (c *WsConn) RemoteAddr() net.Addr {
	return c.conn.UnderlyingConn().RemoteAddr()
}

func (c *WsConn) SetDeadline(t time.Time) error {
	if err := c.conn.UnderlyingConn().SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

func (c *WsConn) SetReadDeadline(t time.Time) error {
	return c.conn.UnderlyingConn().SetReadDeadline(t)
}

func (c *WsConn) SetWriteDeadline(t time.Time) error {
	return c.conn.UnderlyingConn().SetWriteDeadline(t)
}

// ////

type Ws struct {
	Addr string
	Path string
}

func (w Ws) StartServer(h core.Handler) {

	upgrader := websocket.Upgrader{}
	http.HandleFunc(w.Path, func(w http.ResponseWriter, r *http.Request) {

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("err: %v\n",err)
			return
		}

		wsc := WsConn{c, nil}
		h.Handle(&wsc)

	})

	err := http.ListenAndServe(w.Addr, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (w Ws) StartDialer(h core.Handler) {

	u := url.URL{
		Scheme: "ws",
		Host:   w.Addr,
		Path:   w.Path,
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("err: %v\n", err)
		return
	}
	wsc := WsConn{c, nil}
	h.Handle(&wsc)
}

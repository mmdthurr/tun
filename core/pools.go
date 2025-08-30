package core

import (
	"net"
	"sync"
)

type Connection struct {
	mu     sync.Mutex
	Stream net.Conn
}

type Pool struct {
	mu sync.Mutex

	Size int    // for dialer
	Addr string // for dialer
	Tls  bool   // for dialer

	Conns []*Connection
	Smng  *Sessions
}

func (p *Pool) Write(b []byte) {
	for _, c := range p.Conns {
		if c.mu.TryLock() {
			defer c.mu.Unlock()
			_, err := c.Stream.Write(b)
			if err != nil {
				p.Remove(c.Stream)
				continue
			}
			return
		}
	}

}

func NewPool(max_size int) *Pool {
	p := Pool{
		Size: max_size,
	}
	return &p
}

func (p *Pool) Add(conn net.Conn) {

	p.mu.Lock()
	defer p.mu.Unlock()

	c := Connection{
		Stream: conn,
	}

	p.Conns = append(p.Conns, &c)

}

func (p *Pool) Remove(conn net.Conn) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, c := range p.Conns {
		if c.Stream == conn {
			p.Conns = append(p.Conns[:i], p.Conns[i+1:]...)
			conn.Close()
			return true
		}
	}
	return false
}

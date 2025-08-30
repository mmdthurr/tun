package core

import (
	"crypto/tls"
	"io"
	"net"
	"sync"
	"time"
)

type Dialer struct {
	Smng    *Sessions
	Pools   []*Pool
	BckAddr string
}

func ReadOnBck(c net.Conn, id uint16, p *Pool) {
	for {
		buff := make([]byte, 8*4096)
		n, err := c.Read(buff)
		if err != nil {
			cf := Frame{
				Flag:    1,
				Session: id,
				Payload: nil,
			}
			p.Write(cf.Encode())
			return
		}
		f := Frame{
			Flag:    0,
			Session: id,
			Size:    uint16(n),
			Payload: buff[:n],
		}
		p.Write(f.Encode())
	}
}

func (d *Dialer) Dial(addr string, p *Pool) {

	c, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	if p.Tls {
		conf := tls.Config{
			InsecureSkipVerify: true,
		}
		c = tls.Client(c, &conf)
	}

	p.Add(c)
	defer p.Remove(c)

	for {
		buff := make([]byte, 5)
		n, err := c.Read(buff)
		if err != nil {
			return
		}

		f := DecodeFrame(buff[:n])
		switch f.Flag {
		case 0:
			{
				sc := d.Smng.Get(f.Session)

				if sc == nil {
					bckc, err := net.Dial("tcp", d.BckAddr)
					if err != nil {
						continue
					}
					d.Smng.Add(f.Session, bckc)
					sc = d.Smng.Get(f.Session)
					go ReadOnBck(bckc, f.Session, p)
				}
				io.CopyN(sc, c, int64(f.Size))
			}
		case 1:
			{
				d.Smng.Remove(f.Session)
			}
		}
	}
}

func (d *Dialer) StartDialPool(p *Pool) {
	for range p.Size {
		go func() {
			for {
				d.Dial(p.Addr, p)
				time.Sleep(500 * time.Millisecond)
			}
		}()
	}
}

func (d *Dialer) Start() {
	var wg sync.WaitGroup
	wg.Add(1)

	for _, p := range d.Pools {
		go d.StartDialPool(p)
	}

	wg.Wait()
}

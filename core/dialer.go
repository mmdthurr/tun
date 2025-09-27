package core

import (
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/xtaci/smux"
)

type Dialer struct {
	Pools   []*Pool
	BckAddr string

	SmuxConf *smux.Config
}

func (d *Dialer) Dial(addr string, p *Pool) {

	c, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	defer c.Close()

	if p.Tls {
		conf := tls.Config{
			InsecureSkipVerify: true,
		}
		c = tls.Client(c, &conf)
	}

	smuxsrv, err := smux.Server(c, d.SmuxConf)
	if err != nil {
		return
	}
	for {
		stream, err := smuxsrv.AcceptStream()
		if err != nil {
			return
		}

		go func(s *smux.Stream) {
			bck, err := net.Dial("tcp", d.BckAddr)
			if err != nil {
				s.Close()
				return
			}

			Copy(bck, s)

		}(stream)
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

package proto

import (
	"log"
	"net"
	"tun/core"
)

type Tcp struct {
	Addr string
}

func (t Tcp) StartServer(h core.Handler) {

	l, err := net.Listen("tcp", t.Addr)
	if err != nil {
		log.Fatal(err)
		return
	}

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
			return
		}
		go h.Handle(c)
	}

}

func (t Tcp) StartDialer(h core.Handler) {
	c, err := net.Dial("tcp", t.Addr)
	if err != nil {
		log.Printf("err: %v\n", err)
		return
	}
	h.Handle(c)
}

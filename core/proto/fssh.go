// fake ssh dose only perform version exchange part of handshake
// only penetrate firewall
package proto

import (
	"log"
	"net"
	"slices"
	"strings"
	"tun/core"
)

type Fssh struct {
	Addr        string
	Version     string
	TrustedAddr []string
}

func (fs Fssh) StartServer(h core.Handler) {

	l, err := net.Listen("tcp", fs.Addr)
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

		go func() {

			inaddr := strings.Split(c.RemoteAddr().String(), ":")[0]
			if slices.Contains(fs.TrustedAddr, inaddr) {
				_, err := c.Write([]byte(fs.Version))
				if err != nil {
					return
				}

				vb := make([]byte, 200)
				n, err := c.Read(vb)
				if err != nil {
					return
				}

				log.Print(string(vb[:n]))
			}

			h.Handle(c)
		}()
	}

}

func (fs Fssh) StartDialer(h core.Handler) {
	c, err := net.Dial("tcp", fs.Addr)
	if err != nil {
		log.Printf("err: %v\n", err)
		return
	}
	vb := make([]byte, 200)
	n, err := c.Read(vb)
	if err != nil {
		return
	}

	log.Print(string(vb[:n]))

	_, err = c.Write([]byte(fs.Version))
	if err != nil {
		return
	}

	h.Handle(c)
}

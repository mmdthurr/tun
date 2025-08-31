package core

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/xtaci/smux"
)

type TransportSec struct {
	Type       string
	Cert       string
	Key        string
	Servername string
}

type Listener struct {
	Laddr string
	Sec   TransportSec
	Pools map[string]*Pool
}

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

func Copy(c1, c2 net.Conn) {

	defer c1.Close()
	defer c2.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(c1, c2)
	}()

	go func() {
		defer wg.Done()
		io.Copy(c2, c1)
	}()

	wg.Wait()

}

func (li *Listener) Dispatch(c net.Conn) {
	inaddr := strings.Split(c.RemoteAddr().String(), ":")[0]
	log.Println(inaddr)
	p, ok := li.Pools[inaddr]
	if ok {
		// peer
		session, err := smux.Client(c, nil)
		if err != nil {
			return
		}
		p.Add(session)

	} else {

		defer c.Close()

		tmp := make([]byte, 4096)
		n, err := c.Read(tmp)
		if err != nil {
			return
		}

		h, ok := GetHost(tmp[:n])
		if !ok {
			// no host response
			return
		}
		fmt.Println(h)

		p, ok := li.Pools[h]
		if !ok {
			// no pool response
			return
		}
		// test pool
		//p := li.Pools["127.0.0.1"]

		stream := p.OpenStream()
		if stream == nil {
			return
		}

		w := false
		for w == false {
			select {
			case <-time.After(2 * time.Second):
				stream = p.OpenStream()
				if stream == nil {
					return
				}
			default:
				stream.Write(tmp[:n])
				w = true
			}
		}

		Copy(stream, c)

	}
}

func (li *Listener) Start() {

	l, err := net.Listen("tcp", li.Laddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	switch li.Sec.Type {
	case "tls":
		{
			cert, err := tls.LoadX509KeyPair(li.Sec.Cert, li.Sec.Key)
			if err != nil {
				log.Printf("tls err %s", err)
			}
			conf := tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			for {
				c, err := l.Accept()
				if err != nil {
					log.Fatal(err)
					return
				}
				tlsc := tls.Server(c, &conf)
				go li.Dispatch(tlsc)
			}

		}
	default:
		{
			for {
				c, err := l.Accept()
				if err != nil {
					return
				}
				go li.Dispatch(c)
			}
		}
	}

}

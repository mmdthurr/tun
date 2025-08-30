package core

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"strings"
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
	Smng  *Sessions
}

func (li *Listener) ReadLoopOnCli(c net.Conn, tmp []byte, p *Pool) {

	defer c.Close()

	id, ok := li.Smng.New(c)
	if !ok {
		return
	}
	defer li.Smng.Remove(id)

	cf := Frame{
		Flag:    1,
		Session: id,
		Payload: nil,
	}
	defer p.Write(cf.Encode())

	ftmp := Frame{
		Flag:    0,
		Session: id,
		Size:    uint16(len(tmp)),
		Payload: tmp,
	}

	// first write
	p.Write(ftmp.Encode())
	for {
		buff := make([]byte, 8*4096)
		n, err := c.Read(buff)
		if err != nil {
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

func (li *Listener) ReadLoopOnPool(c net.Conn, p *Pool) {
	defer p.Remove(c)

	for {
		buff := make([]byte, 5)
		n, err := c.Read(buff)
		if err != nil {
			return
		}
		// frame decoded
		f := DecodeFrame(buff[:n])

		switch f.Flag {
		case 0:
			{
				sc := li.Smng.Get(f.Session)

				if sc == nil {
					cf := Frame{
						Flag:    1,
						Session: f.Session,
						Payload: nil,
					}
					p.Write(cf.Encode())
					continue
				}

				io.CopyN(sc, c, int64(f.Size))
			}
		case 1:
			{
				li.Smng.Remove(f.Session)
			}
		}
	}
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

func (li *Listener) Dispatch(c net.Conn) {
	inaddr := strings.Split(c.RemoteAddr().String(), ":")[0]
	log.Println(inaddr)
	p, ok := li.Pools[inaddr]
	if ok {
		// peer
		p.Add(c)
		go li.ReadLoopOnPool(c, p)
	} else {

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

		p, ok := li.Pools[h]
		if !ok {

			// no pool response
			return
		}
		// test pool
		//p := li.Pools["127.0.0.1"]

		// handling cli in dispatcher
		go li.ReadLoopOnCli(c, tmp[:n], p)
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

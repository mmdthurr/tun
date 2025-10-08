package core

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/xtaci/smux"
	"github.com/xtls/reality"
)

type TransportSec struct {
	Type string
	//tls
	Cert string
	Key  string

	//utls
	UtlsPk      string //base64 key
	FallBack    string
	Servernames []string
}

type Listener struct {
	Laddr string
	Sec   TransportSec
	Pools map[string]*Pool
	ShId  map[string]string

	Fallback string

	SmuxConf *smux.Config
}

func GetShId(h string) string {
	// customize it based on your domain since my domain is kkdfs.usa.khamenei.ir then [1] would result in usa
	split := strings.Split(h, ".")
	if len(split) < 2 {
		return ""
	}
	return split[1]
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

	var wg sync.WaitGroup
	wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer c1.Close()
		defer wg.Done()

		select {

		case <-ctx.Done():
			return
		default:
			io.Copy(c1, c2)
			cancel()
		}
	}()

	go func() {
		defer c2.Close()
		defer wg.Done()

		select {

		case <-ctx.Done():
			return
		default:
			io.Copy(c2, c1)
			cancel()
		}
	}()

	wg.Wait()

}

func (li *Listener) PerformFallback(c net.Conn, tmp []byte) {
	fc, err := net.Dial("tcp", li.Fallback)
	if err != nil {
		return
	}
	fc.Write(tmp)
	Copy(fc, c)
}

func (li *Listener) Dispatch(c net.Conn) {
	inaddr := strings.Split(c.RemoteAddr().String(), ":")[0]
	log.Println(inaddr)
	p, ok := li.Pools[inaddr]
	if ok {
		// peer
		session, err := smux.Client(c, li.SmuxConf)
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
			li.PerformFallback(c, tmp[:n])
			return
		}

		shid := GetShId(h)
		if shid == "" {
			// no shid
			li.PerformFallback(c, tmp[:n])
			return
		}

		a, ok := li.ShId[shid]
		if !ok {
			// return no valid shid
			li.PerformFallback(c, tmp[:n])
			return
		}

		p, ok := li.Pools[a]
		if !ok {
			// no pool response
			li.PerformFallback(c, tmp[:n])
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
	case "utls":
		{
			//tls conf for node to node
			cert, err := tls.LoadX509KeyPair(li.Sec.Cert, li.Sec.Key)
			if err != nil {
				log.Printf("tls err %s", err)
			}
			tlsconf := tls.Config{
				Certificates: []tls.Certificate{cert},
			}

			//utls for client
			pk, err := base64.RawURLEncoding.DecodeString(li.Sec.UtlsPk)
			if err != nil {
				log.Fatal(err)
				return
			}

			var dialer net.Dialer
			rconf := &reality.Config{
				DialContext: dialer.DialContext,

				Show:                   false,
				Type:                   "tcp",
				Dest:                   li.Sec.FallBack,
				Xver:                   byte(0),
				PrivateKey:             pk,
				MaxTimeDiff:            0,
				NextProtos:             nil, // should be nil
				SessionTicketsDisabled: true,
			}

			rconf.ServerNames = make(map[string]bool)
			for _, sni := range li.Sec.Servernames {
				rconf.ServerNames[sni] = true
			}

			// only empty short ids are accepted
			rconf.ShortIds = make(map[[8]byte]bool)
			var k [8]byte
			rconf.ShortIds[k] = true

			go reality.DetectPostHandshakeRecordsLens(rconf)

			for {
				c, err := l.Accept()
				if err != nil {
					return
				}

				// pre dispatch cause implementing reality client is not feasible right now
				go func() {
					inaddr := strings.Split(c.RemoteAddr().String(), ":")[0]
					_, ok := li.Pools[inaddr]
					if ok {
						tlsc := tls.Server(c, &tlsconf)
						go li.Dispatch(tlsc)

					} else {
						c, err = reality.Server(context.Background(), c, rconf)
						if err != nil {
							return
						}
						go li.Dispatch(c)
					}
				}()
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

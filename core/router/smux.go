package router

import (
	"fmt"
	"net"
	"time"
	"tun/core/utils"

	"github.com/xtaci/smux"
)

type SmuxS2S struct {
	//InAddrPeer []string
	SmuxPool *Pool
	SmuxConf *smux.Config
}

func (ss *SmuxS2S) Route(c net.Conn) {
	session, err := smux.Client(c, ss.SmuxConf)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.Close()
		return
	}

	ss.SmuxPool.Add(session)

}

type SmuxS2C struct {
	SmuxPool     *Pool
	SmuxConf     *smux.Config
	RetryOnWrite int
}

func (ss *SmuxS2C) Route(c net.Conn) {

	defer c.Close()

	buff := make([]byte, 4096)
	n, err := c.Read(buff)
	if err != nil {
		return
	}
	buff = buff[:n]

	stream := ss.SmuxPool.OpenStream()
	if stream == nil {
		return
	}

	i := 0
	for i < ss.RetryOnWrite {
		i++
		select {
		case <-time.After(2 * time.Second):
			stream = ss.SmuxPool.OpenStream()
			if stream == nil {
				return
			}
		default:
			_, err := stream.Write(buff)
			if err != nil {
				return
			}
			i = ss.RetryOnWrite
		}
	}

	utils.Copy(stream, c)
}

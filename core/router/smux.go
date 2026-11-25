package router

import (
	"context"
	"fmt"
	"net"
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

	stream := ss.SmuxPool.OpenStream(context.Background())
	if stream == nil {
		return
	}

	var buff [4096]byte
	//buff := make([]byte, 4096)
	n, err := c.Read(buff[:])
	if err != nil {
		return
	}
	readbuff := buff[:n]

	// this is buggy
	i := 0
	for i <= ss.RetryOnWrite {
		i++
		stream = ss.SmuxPool.OpenStream(context.Background())
		if stream == nil {
			return
		}
		_, err := stream.Write(readbuff)
		if err != nil {
			return
		} else {
			break
		}

	}

	utils.Copy(stream, c)
}

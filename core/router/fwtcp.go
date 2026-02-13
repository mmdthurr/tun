package router

import (
	"log"
	"net"
	"tun/core/utils"
)

type ForwarderTcp struct {
	//InAddrPeer []string
	BackAddr string
}

func (f ForwarderTcp) Route(c net.Conn) {
	defer c.Close()

	bc, err := net.Dial("tcp", f.BackAddr)
	if err != nil {
		log.Println(err)
		return
	}

	utils.Copy(bc, c)
}

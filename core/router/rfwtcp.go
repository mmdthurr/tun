package router

import (
	"net"
	"sync/atomic"
	"tun/core/utils"
)

type RForwarderTcp struct {
	next      uint32
	BackAddrs []string
}

func (f RForwarderTcp) Route(c net.Conn) {
	defer c.Close()

	if len(f.BackAddrs) == 0 {
		return
	}

	n := atomic.AddUint32(&f.next, 1)
	if int(n) > len(f.BackAddrs) {
		atomic.StoreUint32(&f.next, 1)
		n = 1
	}

	bc, err := net.Dial("tcp", f.BackAddrs[(int(n)-1)%len(f.BackAddrs)])
	if err != nil {
		return
	}

	utils.Copy(bc, c)
}

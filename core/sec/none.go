package sec

import (
	"net"
)

type None struct{}

func (n None) WrapConn(c net.Conn) net.Conn {
	return c
}

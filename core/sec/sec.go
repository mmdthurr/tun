package sec

import (
	"net"
)

type Sec interface {
	WrapConn(c net.Conn) net.Conn
}

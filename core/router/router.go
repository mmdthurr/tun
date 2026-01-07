package router

import "net"

type Router interface {
	Route(c net.Conn)
}

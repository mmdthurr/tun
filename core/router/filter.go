package router

import (
	"bytes"
	"io"
	"log"
	"net"
	"strings"
	"tun/core/utils"
)

type Filter interface {
	// filter return specific router tag
	GetRouter(c net.Conn) (string, net.Conn)
}

type DecoyFilter struct {
	DefaultRouter string
}

func (df *DecoyFilter) GetRouter(c net.Conn) (string, net.Conn) {
	return df.DefaultRouter, c
}

// host header
type FilterHostHeader struct {
	//hostheader to router tag
	HostHeaderMap map[string]string
	DefaultRouter string
}

func (fhh *FilterHostHeader) GetRouter(c net.Conn) (string, net.Conn) {

	buf := make([]byte, 4096)
	n, err := c.Read(buf)
	if err != nil {
		return "", nil
	}

	r := io.MultiReader(bytes.NewReader(buf[:n]), c)
	new_c := utils.WrapperConn{
		Conn:   c,
		Reader: r,
	}

	h, ok := utils.GetHost(buf)
	if !ok {
		log.Printf("%s - %s\n", c.RemoteAddr(), "no host header")
		return fhh.DefaultRouter, new_c
	}


	log.Printf("%s - %s\n", c.RemoteAddr(), h)

	t, ok := fhh.HostHeaderMap[h]
	if !ok {
		return fhh.DefaultRouter, new_c
	}

	return t, new_c
}

type FilterAddr struct {
	//ip to router tag map
	AddrMap map[string]string

	// set it to next inorder to pass
	DefaultRouter string
}

func (fa *FilterAddr) GetRouter(c net.Conn) (string, net.Conn) {

	inaddr := strings.Split(c.RemoteAddr().String(), ":")[0]
	tag, ok := fa.AddrMap[inaddr]
	if ok {
		log.Printf("%s - %s\n", inaddr, tag)
		return tag, c
	}

	log.Printf("%s - %s\n", inaddr, "default tag")
	return fa.DefaultRouter, c

}

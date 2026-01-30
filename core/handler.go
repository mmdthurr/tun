package core

import (
	"log"
	"net"
	"tun/core/router"
	"tun/core/sec"

	"github.com/xtaci/smux"
)

type Handler interface {
	Handle(net.Conn)
}

type HandlerSmuxDialer struct {
	ConSec sec.Sec
	// handler is not flexible
	// for now only limited to smux
	SmuxConf *smux.Config
	Router   router.Router
}

func (h *HandlerSmuxDialer) Handle(c net.Conn) {

	c = h.ConSec.WrapConn(c)
	if c == nil {
		return
	}

	smuxsrv, err := smux.Server(c, h.SmuxConf)
	if err != nil {
		log.Fatal(err)
		return
	}
	for {
		stream, err := smuxsrv.AcceptStream()
		if err != nil {
			return
		}
		go h.Router.Route(stream)
	}

}

type HandlerServer struct {
	ConSec      sec.Sec
	FiltersList []router.Filter
	//FilterAddr       router.Filter
	//FilterHostHeader router.Filter
	//tag to router
	MapRouter map[string]router.Router
}

func (h *HandlerServer) Handle(c net.Conn) {

	c = h.ConSec.WrapConn(c)
	if c == nil {
		return
	}

	t := ""
	for _, f := range h.FiltersList {
		t, c = f.GetRouter(c)
		//reserved tag name is next
		if t == "next" {
			continue
		} else {
			break
		}
	}

	r, ok := h.MapRouter[t]
	if ok {
		r.Route(c)
	}
}

package sec

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"tun/core/utils"
)

type MultiplexServer struct {
	SniMap  map[string]Sec
	Default Sec
}

// https://github.com/sausagenoods/snitch/blob/master/conn.go
func readClientHello(reader io.Reader) (*tls.ClientHelloInfo, error) {
	var hello *tls.ClientHelloInfo
	err := tls.Server(utils.MockConn{Reader: reader}, &tls.Config{
		GetConfigForClient: func(argHello *tls.ClientHelloInfo) (*tls.Config, error) {
			hello = new(tls.ClientHelloInfo)
			*hello = *argHello
			return nil, nil
		},
	}).Handshake()
	if hello == nil {
		return nil, err
	}
	return hello, nil
}

func peekClientHello(reader io.Reader) (*tls.ClientHelloInfo, io.Reader, error) {
	peekedBytes := new(bytes.Buffer)
	hello, err := readClientHello(io.TeeReader(reader, peekedBytes))
	if err != nil {
		return nil, io.MultiReader(peekedBytes, reader), err
	}
	return hello, io.MultiReader(peekedBytes, reader), nil
}
//------------

func (f *MultiplexServer) WrapConn(c net.Conn) net.Conn {

	clientHello, r, err := peekClientHello(c)
	c = utils.WrapperConn{
		Conn:   c,
		Reader: r,
	}

	if err != nil {
		return f.Default.WrapConn(c)
	}

	sec, ok := f.SniMap[clientHello.ServerName]
	if ok {
		return sec.WrapConn(c)
	}

	return f.Default.WrapConn(c)

}

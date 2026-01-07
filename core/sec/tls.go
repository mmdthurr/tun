package sec

import (
	"crypto/tls"
	"log"
	"net"
)

type TlsServer struct {
	Config *tls.Config
}

func (t TlsServer) WrapConn(c net.Conn) net.Conn {
	return tls.Server(c, t.Config)
}

type TlsClient struct {
	Config *tls.Config
}

func (t TlsClient) WrapConn(c net.Conn) net.Conn {
	return tls.Client(c, t.Config)
}

func BuildTlsClient(allowinsecure bool, servername string) TlsClient {

	conf := tls.Config{
		InsecureSkipVerify: allowinsecure,
		ServerName:         servername,
	}

	return TlsClient{
		&conf,
	}
}
func BuildTlsServer(key_path string, cert_path string) TlsServer {

	cert, err := tls.LoadX509KeyPair(cert_path, key_path)
	if err != nil {
		log.Printf("err: %v\n", err)
	}
	conf := tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return TlsServer{
		&conf,
	}
}

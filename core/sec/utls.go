package sec

import (
	"context"
	"encoding/base64"
	"log"
	"net"

	"github.com/xtls/reality"
)

type UtlsServer struct {
	Conf *reality.Config
}

func (ut UtlsServer) WrapConn(c net.Conn) net.Conn {

	rc, err := reality.Server(context.Background(), c, ut.Conf)
	if err != nil {
		log.Printf("err: %v\n", err)
		return nil
	}
	return rc

}

func BuildUtls(pk_ string, falb string, servernames []string) UtlsServer {

	//utls for client
	pk, err := base64.RawURLEncoding.DecodeString(pk_)
	if err != nil {
		log.Fatal(err)
	}

	var dialer net.Dialer
	rconf := &reality.Config{
		DialContext: dialer.DialContext,

		Show:                   false,
		Type:                   "tcp",
		Dest:                   falb,
		Xver:                   byte(0),
		PrivateKey:             pk,
		MaxTimeDiff:            0,
		NextProtos:             nil, // should be nil
		SessionTicketsDisabled: true,
	}
	rconf.ServerNames = make(map[string]bool)
	for _, sni := range servernames {
		rconf.ServerNames[sni] = true
	}

	// only empty short ids are accepted
	rconf.ShortIds = make(map[[8]byte]bool)
	var k [8]byte
	rconf.ShortIds[k] = true

	go reality.DetectPostHandshakeRecordsLens(rconf)

	return UtlsServer{
		Conf: rconf,
	}
}

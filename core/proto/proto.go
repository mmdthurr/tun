package proto

import "tun/core"

type Proto interface {
	StartDialer(core.Handler)
	StartServer(core.Handler)
}

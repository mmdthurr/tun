package main

import (
	"flag"
	"mmd/tun/core"
	"net"
	"sync"
)

func main() {
	configpath := flag.String("c", "/etc/tun/config.json", "this is the config path")
	flag.Parse()

	conf := core.GetConfig(*configpath)

	switch conf.Mode {
	case "listener":
		{
			p := core.NewPool(0)
			smng := &core.Sessions{
				Sc: make(map[uint16]net.Conn),
			}
			p.Smng = smng

			peers := make(map[string]*core.Pool)
			for _, peer := range conf.TrustedPeers {
				peers[peer] = core.NewPool(0)
			}

			if conf.Sec == "tls" {
				l := core.Listener{
					Laddr: conf.Laddr,
					Pools: peers,
					Smng:  smng,
					Sec: core.TransportSec{
						Type: "tls",
						Key:  conf.TlsSetting.Key,
						Cert: conf.TlsSetting.Cert,
					},
				}

				l.Start()
			} else {

				l := core.Listener{
					Laddr: conf.Laddr,
					Pools: peers,
					Smng:  smng,
					Sec: core.TransportSec{
						Type: "",
					},
				}

				l.Start()
			}
		}

	case "dialer":
		{
			var wg sync.WaitGroup
			wg.Add(1)

			smng := &core.Sessions{
				Sc: make(map[uint16]net.Conn),
			}
			var pools []*core.Pool
			for _, peer := range conf.Peers {

				p := core.NewPool(peer.PoolSize)
				p.Addr = peer.Addr
				p.Tls = peer.Tls
				p.Smng = smng
				pools = append(pools, p)

			}
			d := core.Dialer{
				Smng:    smng,
				Pools:   pools,
				BckAddr: conf.BckAddr,
			}
			d.Start()

			wg.Wait()
		}
	}

}

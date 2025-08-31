package main

import (
	"flag"
	"mmd/tun/core"
	"sync"
)

func main() {
	configpath := flag.String("c", "/etc/tun/config.json", "this is the config path")
	flag.Parse()

	conf := core.GetConfig(*configpath)

	switch conf.Mode {
	case "listener":
		{
			peers := make(map[string]*core.Pool)
			for _, peer := range conf.TrustedPeers {
				peers[peer] = core.NewPool(0)
			}

			if conf.Sec == "tls" {
				l := core.Listener{
					Laddr: conf.Laddr,
					Pools: peers,
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

			var pools []*core.Pool
			for _, peer := range conf.Peers {

				p := core.NewPool(peer.PoolSize)
				p.Addr = peer.Addr
				p.Tls = peer.Tls
				pools = append(pools, p)

			}
			d := core.Dialer{
				Pools:   pools,
				BckAddr: conf.BckAddr,
			}
			d.Start()

			wg.Wait()
		}
	}

}

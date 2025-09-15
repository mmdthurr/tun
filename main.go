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
			shids := make(map[string]string)

			for _, peer := range conf.TrustedPeers {
				peers[peer.Addr] = core.NewPool(0)
				shids[peer.ShortID] = peer.Addr
			}

			switch conf.Sec {

			case "tls":
				{
					l := core.Listener{
						Laddr: conf.Laddr,
						Pools: peers,
						ShId:  shids,
						Sec: core.TransportSec{
							Type: "tls",
							Key:  conf.TlsSetting.Key,
							Cert: conf.TlsSetting.Cert,
						},
						Fallback: conf.FallBack,
					}

					l.Start()
				}
			case "utls":
				{
					l := core.Listener{
						Laddr: conf.Laddr,
						Pools: peers,
						ShId:  shids,
						Sec: core.TransportSec{
							Type: "utls",

							//tls
							Key:  conf.TlsSetting.Key,
							Cert: conf.TlsSetting.Cert,

							//utls
							UtlsPk:      conf.UtlsSetting.PrivateKey,
							FallBack:    conf.UtlsSetting.FallBack,
							Servernames: conf.UtlsSetting.ServerNames,
						},
						Fallback: conf.FallBack,
					}

					l.Start()
				}
			default:
				{
					l := core.Listener{
						Laddr: conf.Laddr,
						Pools: peers,
						Sec: core.TransportSec{
							Type: "",
						},
						Fallback: conf.FallBack,
					}

					l.Start()

				}
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

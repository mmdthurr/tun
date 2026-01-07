package context

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"tun/core"
	"tun/core/proto"
	"tun/core/router"
	"tun/core/sec"

	"github.com/xtaci/smux"
)

func build_tls_utls_server(s core.SecConf) sec.Sec {

	switch s.SecMode {
	case "tls-server":
		{
			return sec.BuildTlsServer(
				s.Key,
				s.Cert,
			)
		}
	case "utls":
		{
			return sec.BuildUtls(
				s.Pk,
				s.Fallback,
				s.ServerNames,
			)
		}
	default:
		return sec.None{}
	}

}

func BuildFromConf(config core.Conf) {

	routerMap := make(map[string]router.Router)
	for t, r := range config.RouterMap {
		switch r.Mode {
		case "fwtcp":
			{
				routerMap[t] = &router.ForwarderTcp{
					BackAddr: r.BackAddr,
				}
			}
		//smux mode will result in two tag 1.tag 2.tagpool set filter addr
		case "smux":
			{
				pool := router.NewPool(0)
				smuxconf := smux.DefaultConfig()

				tpool := fmt.Sprintf("%spool", t)
				routerMap[tpool] = &router.SmuxS2S{
					SmuxPool: pool,
					SmuxConf: smuxconf,
				}
				routerMap[t] = &router.SmuxS2C{

					SmuxPool:     pool,
					SmuxConf:     smuxconf,
					RetryOnWrite: r.RetryOnWrite,
				}

			}
		}

	}

	handlerMap := make(map[string]core.Handler)
	for t, h := range config.HandlerMap {

		var security sec.Sec
		security = sec.None{}
		switch h.SecObj.SecMode {
		case "tls-client":
			{
				security = sec.BuildTlsClient(h.SecObj.AllowInsecure, h.SecObj.ServerName)
			}
		case "tls-server":
			{
				security = build_tls_utls_server(h.SecObj)
			}
		case "utls":
			{
				security = build_tls_utls_server(h.SecObj)
			}
		case "multiplex":
			{
				snimap := make(map[string]sec.Sec)
				for sni, obj := range h.SecObj.SniMapp {
					snimap[sni] = build_tls_utls_server(*obj)
				}

				security = &sec.MultiplexServer{
					Default: build_tls_utls_server(*h.SecObj.DefaultSec),
					SniMap:  snimap,
				}
			}
		}

		var filList []router.Filter
		for _, f := range h.Filters {

			switch f.Mode {
			case "addr":
				{
					fobj := router.FilterAddr{
						AddrMap:       f.Map,
						DefaultRouter: f.DefaultTag,
					}
					filList = append(filList, &fobj)
				}
			case "hostheader":
				{
					fobj := router.FilterHostHeader{
						HostHeaderMap: f.Map,
						DefaultRouter: f.DefaultTag,
					}
					filList = append(filList, &fobj)
				}
			case "decoy":
				{
					fobj := router.DecoyFilter{
						DefaultRouter: f.DefaultTag,
					}
					filList = append(filList, &fobj)
				}
			}
		}

		handlerRouterMap := make(map[string]router.Router)
		for _, rt := range h.RouterList {
			handlerRouterMap[rt] = routerMap[rt]
		}

		switch h.Mode {
		case "dialersmux":
			{
				handlerMap[t] = &core.HandlerSmuxDialer{
					ConSec:   security,
					SmuxConf: smux.DefaultConfig(),
					Router:   handlerRouterMap[h.RouterList[0]],
				}
			}
		case "serversmux":
			{

				handlerMap[t] = &core.HandlerServer{
					ConSec:      security,
					FiltersList: filList,
					MapRouter:   handlerRouterMap,
				}

			}

		}

	}

	var wg sync.WaitGroup
	wg.Add(1)
	for _, l := range config.Listeners {

		switch l.NetworkType {
		case "tcp":
			{
				// no listener context directly start it off
				go proto.Tcp{
					Addr: l.Addr,
				}.StartServer(handlerMap[l.HandlerTag])
			}
		case "ws":
			{
				// no listener context directly start it off
				go proto.Ws{
					Addr: l.Addr,
					Path: l.WsPath,
				}.StartServer(handlerMap[l.HandlerTag])
			}
		default:
			log.Printf("not valid network %s \n", l.NetworkType)
		}

	}

	for _, d := range config.Dialers {

		switch d.NetworkType {
		case "tcp":
			{
				go DialerContext{
					DialPool: d.DialSize,
					Dialer: proto.Tcp{
						Addr: d.Addr,
					},
					Handler:     handlerMap[d.HandlerTag],
					TriggerChan: make(chan DialerEvent),
				}.Start()

			}
		case "ws":
			{
				go DialerContext{
					DialPool: d.DialSize,
					Dialer: proto.Ws{
						Addr: d.Addr,
						Path: d.WsPath,
					},
					Handler:     handlerMap[d.HandlerTag],
					TriggerChan: make(chan DialerEvent),
				}.Start()

			}
		default:
			log.Printf("not valid network %s \n", d.NetworkType)
		}

	}
	wg.Wait()

}

func GetConfig(path string) core.Conf {
	configFile, err := os.Open(path)
	if err != nil {
		log.Println("config parse err")
		log.Fatal(err)
	}
	defer configFile.Close()

	config := core.Conf{}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config

}

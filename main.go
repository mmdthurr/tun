package main

import (
	"flag"
	"tun/core/context"
)

func main() {

	conf_path := flag.String("c", "/etc/tun/config.json", "conf path")
	flag.Parse()

	//conf := context.GetConfig("/home/mmd/code/tun/example/test.json")
	conf := context.GetConfig(*conf_path)
	context.BuildFromConf(conf)
}

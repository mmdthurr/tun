package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Tls struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}
type Peer struct {
	Addr     string `json:"addr"`
	Tls      bool   `json:"tls"`
	ShortID  string `json:"shortid"`
	PoolSize int    `json:"poolsize"`
}

type Config struct {
	Mode string `json:"mode"`
	// listener
	Laddr        string `json:"laddr"`
	Sec          string `json:"sec"`
	TlsSetting   Tls    `json:"tls"`
	TrustedPeers []Peer `json:"trustedpeers"`
	FallBack     string `json:"fallback"`
	// dialer
	BckAddr string `json:"bckaddr"`
	Peers   []Peer `json:"peers"`
}

func GetConfig(path string) Config {
	configFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer configFile.Close()

	config := Config{}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config

}

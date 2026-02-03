package core

type ListenerConf struct {
	Addr string `json:"addr"`
	// tcp, ws, fssh
	NetworkType string `json:"network_type"`
	HandlerTag  string `json:"handler_tag"`

	//ws only
	WsPath string `json:"ws_path"`
	//fssh only
	FsshVersion string `json:"fssh_version"`
}

type DialerConf struct {
	Addr        string `json:"addr"`
	NetworkType string `json:"network_type"`
	HandlerTag  string `json:"handler_tag"`
	DialSize    int    `json:"dial_size"`

	//path ws only option
	WsPath string `json:"ws_path"`
	//fssh only
	FsshVersion string `json:"fssh_version"`
}

type FilterConf struct {
	DefaultTag string `json:"default_tag"`

	// addr, hostheader, decoy
	Mode string            `json:"mode"`
	Map  map[string]string `json:"map"`
}

type RouterConf struct {
	// for further smux router configuration rebuild
	// with desired option
	// fwtcp, smux, rfwtcp
	Mode string `json:"mode"`
	// for fwtcp
	BackAddr string `json:"back_addr"`
	// for rfwtcp
	BackAddrs []string `json:"back_addrs"`
	// for smux
	RetryOnWrite int `json:"retry_on_write"`
}

type SecConf struct {
	// tls-server tls-client ,utls, multiplex(sni centered), chacha(symetric), empty for none
	SecMode string `json:"sec_mode"`

	//chacha key 32bit string
	ChaChaKey string `json:"chacha_key"`

	//tls-server settign
	Key  string `json:"key"`
	Cert string `json:"cert"`

	//tls-client setting
	AllowInsecure bool   `json:"allow_insecure"`
	ServerName    string `json:"server_name"`

	//utls
	Pk          string   `json:"pk"`
	Fallback    string   `json:"fall_back"`
	ServerNames []string `json:"server_names"`

	//multiplexSetting
	// only valid tls ,utls
	DefaultSec *SecConf            `json:"default_sec"`
	SniMapp    map[string]*SecConf `json:"sni_map"`
}

type HandlerConf struct {

	// dialersmux, serversmux
	Mode string `json:"mode"`

	SecObj SecConf `json:"sec"`

	// Fifo like filter list but filters will end
	// if toppest filter return tag it won't call
	// other filter.
	// on err on each filter return "" tag.
	// if default tag set to "next" then the next
	// filter will be applied on the conn.
	Filters []FilterConf `json:"filters"`

	// list of assigned map to this
	RouterList []string `json:"routers"`
}

type Conf struct {
	RouterMap  map[string]RouterConf  `json:"router_map"`
	HandlerMap map[string]HandlerConf `json:"handler_map"`
	Listeners  []ListenerConf         `json:"listeners"`
	Dialers    []DialerConf           `json:"dialers"`
}

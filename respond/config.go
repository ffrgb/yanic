package respond

import "github.com/FreifunkBremen/yanic/lib/duration"

type Config struct {
	Enable          bool              `toml:"enable"`
	Synchronize     duration.Duration `toml:"synchronize"`
	Interfaces      []InterfaceConfig `toml:"interfaces"`
	Sites           []string          `toml:"sites"`
	Port            int               `toml:"port"`
	CollectInterval duration.Duration `toml:"collect_interval"`
}

type InterfaceConfig struct {
	InterfaceName  string `toml:"ifname"`
	Port           int    `toml:"port"`
	IP             string `toml:"ip"`
	MulticastGroup string `toml:"multicast"`
}

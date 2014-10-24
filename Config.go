package gnlog

import (
	gnet "../gnet"
)

type Config struct {
	LogDir			string		`json:"log_dir"`
	DataDir			string		`json:"data_dir"`
	Server			gnet.Config	`json:"server"`
}

var (
	Conf Config
)

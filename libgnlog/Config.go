package gnlog

import (
	"github.com/rare/gnet"
)

type AuthConfig struct {
	Enable			bool		`json:"enable"`
	Filename		string		`json:"filename"`
}

type GNLogConfig struct {
	LogDir			string		`json:"log_dir"`
	DataDir			string		`json:"data_dir"`
	Auth			AuthConfig	`json:"auth"`
	Server			gnet.Config	`json:"server"`
}

var (
	Conf GNLogConfig
)

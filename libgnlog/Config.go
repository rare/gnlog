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
	LogChanBufSize	uint32		`json:"log_chan_buf_size"`
	MaxLogFileSize	int64		`json:"max_log_file_size"`
	Auth			AuthConfig	`json:"auth"`
	Server			gnet.Config	`json:"server"`
}

var (
	Conf GNLogConfig
)

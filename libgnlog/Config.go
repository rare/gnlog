package gnlog

import (
	"github.com/rare/gnet"
)

type AuthConfig struct {
	Enable			bool		`json:"enable"`
	Filename		string		`json:"filename"`
}

type ByTimePolicy struct {
	Enable			bool		`json:"enable"`
	Rule			string		`json:"rule"`
}

type BySizePolicy struct {
	Enable			bool		`json:"enable"`
	MaxLogFileSize	int64		`json:"max_log_file_size"`
}

type SplitPolicy struct {
	ByTime			ByTimePolicy	`json:"by_time"`
	BySize			BySizePolicy	`json:"by_size"`
}

type LogConfig struct {
	Dir				string		`json:"dir"`
	BufSize			uint32		`json:"buf_size"`
	SplitPolicy		SplitPolicy	`json:"split_policy"`
}

type GNLogConfig struct {
	Auth			AuthConfig	`json:"auth"`
	Log				LogConfig	`json:"log"`
	Server			gnet.Config	`json:"server"`
}

var (
	Conf GNLogConfig
)

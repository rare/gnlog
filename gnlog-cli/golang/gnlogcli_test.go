package gnlogcli

import (
	"fmt"
	"os"
	"testing"
	"time"
	log "github.com/cihub/seelog"
)

var cli *GNLogClient = NewGNLogClient()

func init() {
	var conf Config = Config {
		Addr:		"localhost:20000",
		Auth:		"abcdefg",
		Catalog:	"www.123.com",
		Filename:	"user.log",
		Minlevel:	log.InfoLvl,
		Format:		"%Date/%Time [%LEV] %Msg%n",
	}
	if err := Init(&conf); err != nil {
		fmt.Printf("init log error: (%v)\n", err)
		os.Exit(1)
	}
}

func TestLog(t *testing.T) {
	for {
		log.Trace("trace message")
		log.Debug("debug message")
		log.Info("info message")
		log.Warn("warn message")
		log.Error("error message")
		time.Sleep(time.Second)
	}
}


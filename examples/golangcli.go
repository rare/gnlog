package main

import (
	"fmt"
	"os"
	"time"
	"github.com/rare/gnlog/gnlog-cli/golang"
	log "github.com/cihub/seelog"
)

func init() {
	var conf gnlogcli.Config = gnlogcli.Config {
		Addr:		"localhost:20000",
		Auth:		"abcdefg",
		Catalog:	"www.123.com",
		Filename:	"user.log",
		Minlevel:	log.InfoLvl,
		Format:		"%Date/%Time [%LEV] %Msg%n",
	}
	if err := gnlogcli.Init(&conf); err != nil {
		fmt.Printf("init log error: (%v)\n", err)
		os.Exit(1)
	}
}

func main() {
	for {
		log.Trace("trace message")
		log.Debug("debug message")
		log.Info("info message")
		log.Warn("warn message")
		log.Error("error message")
		time.Sleep(time.Second)
	}
}


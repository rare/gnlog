package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rare/gnet"
	"github.com/rare/gnlog/libgnlog"
	"github.com/rare/gnlog/conf"
	log "github.com/cihub/seelog"
)

//globals definition
var (
	CONF_PATH = flag.String("c", "./conf/conf.json", "conf file path")
)

func initSignal(svr *gnet.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		svr.Stop()
	}()
}

func loadConfig() {
	err := conf.LoadConfig(*CONF_PATH, &gnlog.Conf)
	if err != nil {
		fmt.Printf("LoadConfig (%s) failed: (%v)", *CONF_PATH, err)
		os.Exit(1)
	}
}

func initLog() {
	logger, err := log.LoggerFromConfigAsFile("./conf/log.xml")
	if err != nil {
		fmt.Printf("Load log config failed: (%v)", err)
		os.Exit(1)
	}
	log.ReplaceLogger(logger)
}

//init function
func init() {
	flag.Parse()
	loadConfig()
	initLog()
}

func initAuth() {
	if err := gnlog.Auth.Init(); err != nil {
		log.Criticalf("Auth.Init error: (%v)", err)
	}
}

func initHandler(svr *gnet.Server) {
	svr.HandleFunc(gnlog.CMD_START, gnlog.HandleStart)
	svr.HandleFunc(gnlog.CMD_LOG, gnlog.HandleLog)
}

func main() {
	wg := &sync.WaitGroup{}
	svr := gnet.NewServer()

	initSignal(svr)

	err := svr.Init(&gnlog.Conf.Server)
	if err != nil {
		log.Criticalf("Init Server error: (%v)", err)
		return
	}

	initAuth()
	initHandler(svr)

	wg.Add(1)
	go func(){
		svr.Run()
		wg.Done()
	}()
	wg.Wait()
}

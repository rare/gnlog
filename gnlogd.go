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
)

//globals definition
var (
	CONF_PATH = flag.String("c", "./conf/conf.json", "conf file path")
)

func initSignal(wg *sync.WaitGroup) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		wg.Done()
	}()
}

func loadConfig() {
	err := conf.LoadConfig(*CONF_PATH, &gnlog.Conf)
	if err != nil {
		fmt.Printf("LoadConfig (%s) failed: (%s).\n", *CONF_PATH, err)
		os.Exit(1)
	}
}

func initLog() {
	gnlog.InitLog()
}

//init function
func init() {
	flag.Parse()
	loadConfig()
	initLog()
}

func initHandler(svr *gnet.Server) {
	svr.HandleFunc(gnlog.CMD_START, gnlog.HandleStart)
	svr.HandleFunc(gnlog.CMD_LOG, gnlog.HandleLog)
}

func main() {
	wg := &sync.WaitGroup{}

	initSignal(wg)

	svr := gnet.NewServer()
	err := svr.Init(&gnlog.Conf.Server)
	if err != nil {
		fmt.Printf("Init Server error: (%s).\n", err)
		return
	}
	
	initHandler(svr)

	wg.Add(1)
	go func(){
		svr.Run()
		wg.Done()
	}()
	wg.Wait()
}

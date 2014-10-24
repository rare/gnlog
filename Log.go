package gnlog

import (
	"log"
	"os"
)

func InitLog() {
	//init logger
}

var (
	logger = log.New(os.Stdout, "gnlog: ", log.Lshortfile)
)

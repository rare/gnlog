package gnlog

import (
	"log"
	"os"
)

var (
)

func InitLog() {

	//init logger
}

var (
	logger = log.New(os.Stdout, "gnlog: ", log.Lshortfile)
)

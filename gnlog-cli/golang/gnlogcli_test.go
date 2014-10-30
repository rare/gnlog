package gnlogcli

import (
	"fmt"
	"testing"
	"os"
)

var cli *GNLogClient = NewGNLogClient()

func init() {
	if err := cli.Init("localhost:20000", "abcdefg", "www.123.com", "user.log"); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func TestLog(t *testing.T) {
	for {
		err := cli.Log("test")
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}


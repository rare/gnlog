package gnlog

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

type Authorizer struct {
	auth_list	map[string]bool
}

func NewAutorizer() *Authorizer {
	return &Authorizer {
		auth_list: make(map[string]bool),
	}
}

func (this *Authorizer) Init() error {
	file, err := os.Open(Conf.Auth.Filename)
	if err != nil {
		//TODO
		return errors.New("open auth file error")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		auth := strings.Trim(scanner.Text(), " \t")
		if auth != "" {
			this.auth_list[auth] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.New("read auth file error")
	}
	return nil
}

func (this *Authorizer) Check(auth string) bool {
	if !Conf.Auth.Enable {
		return true
	}

	flag, ok := this.auth_list[auth]
	return ok && flag
}

var (
	Auth = NewAutorizer()
)

package gnlog

import (
	"bufio"
	"errors"
	"os"
	"strings"
	log "github.com/cihub/seelog"
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
		log.Warnf("open auth file(%s) error: (%v)", Conf.Auth.Filename, err)
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
		log.Warnf("read auth file(%s) error: (%v)", Conf.Auth.Filename, err)
		return errors.New("read auth file error")
	}
	return nil
}

func (this *Authorizer) Check(auth string) error {
	if !Conf.Auth.Enable {
		return nil
	}

	if _, ok := this.auth_list[auth]; !ok {
		return errors.New("auth failed")
	}

	return nil
}

var (
	Auth = NewAutorizer()
)

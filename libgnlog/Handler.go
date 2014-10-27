package gnlog

import (
	"errors"
	"github.com/rare/gnet"
)

func HandleStart(req *gnet.Request, resp *gnet.Response) error {
	buf, err := req.Body()
	if err != nil {
		return err
	}

	cmd := new(StartCmd)
	if cmd.Deserialize(buf) != nil {
		return err
	}

	if !Auth.Check(cmd.Auth) {
		return errors.New("auth error.")
	}

	return nil
}

func HandleLog(req *gnet.Request, resp *gnet.Response) error {
	buf, err := req.Body()
	if err != nil {
		return err
	}

	req.BodyLen()

	//TODO
	return nil
}

package gnlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/rare/gnet"
)

func validateDirAndFilename(catalog string, filename string) (string, string, error) {
	if len(catalog) > MAX_DIR_NAME_LEN {
		return "", "", errors.New("dir name too long")
	}
	if len(filename) > MAX_FILE_NAME_LEN {
		return "", "", errors.New("file name too long")
	}
	catalog = strings.Replace(catalog, strconv.QuoteRune(filepath.Separator), "_", -1)
	filename = strings.Replace(filename, strconv.QuoteRune(filepath.Separator), "_", -1)
	return catalog, filename, nil
}

func HandleStart(req *gnet.Request, resp *gnet.Response) error {
	buf, err := req.Body()
	if err != nil {
		return err
	}

	cmd := new(StartCmd)
	if err := json.Unmarshal(buf, cmd); err != nil {
		return err
	}

	var status int32 = STATUS_OK
	if err := Auth.Check(cmd.Auth); err != nil {
		status = STATUS_AUTH_ERR
		resp.SetCloseAfterSending()
	}

	var srbuf []byte
	sr := new(StartResp)
	sr.Status = status
	srbuf, err = json.Marshal(sr)
	if err != nil {
		return err
	}
	resp.SetBody(bytes.NewBuffer(srbuf))

	req.Client().Storage().Set("mode", cmd.Mode)
	req.Client().Storage().Set("authed", true)

	return nil
}

func HandleLog(req *gnet.Request, resp *gnet.Response) error {
	if _, ok := req.Client().Storage().Get("authed"); !ok {
		return errors.New("channel not authed")
	}

	buf, err := req.Body()
	if err != nil {
		return err
	}

	cmd := new(LogCmd)
	if err := json.Unmarshal(buf, cmd); err != nil {
		return err
	}

	catalog, filename, err := validateDirAndFilename(cmd.Catalog, cmd.Filename)
	if err != nil {
		return err
	}

	logwriter := NewFileLogWriter()
	err = logwriter.Init(catalog, filename)
	if err != nil {
		return err
	}
	logwriter.Write(buf)

	return nil
}

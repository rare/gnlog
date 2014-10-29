package gnlog

import (
	"fmt"			//debug

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
	//debug
	fmt.Println("handle start cmd")

	buf, err := req.Body()
	if err != nil {
		//debug
		fmt.Println("read start cmd body error")
		return err
	}

	cmd := new(StartCmd)
	if err := json.Unmarshal(buf, cmd); err != nil {
		//debug
		fmt.Println("parse start cmd body error")
		return err
	}

	var status int32 = STATUS_OK
	if err := Auth.Check(cmd.Auth); err != nil {
		//debug
		fmt.Println("check auth error")
		status = STATUS_AUTH_ERR
		resp.SetCloseAfterSending()
	}

	var srbuf []byte
	sr := new(StartResp)
	sr.Status = status
	srbuf, err = json.Marshal(sr)
	if err != nil {
		//debug
		fmt.Println("encode resp error")
		return err
	}
	resp.SetBodyLen((uint32)(len(srbuf)))
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
	var mode uint8 = MODE_LINE
	if m, ok := req.Client().Storage().Get("mode"); ok {
		mode = m.(uint8)
	}
	logdata := bytes.NewBufferString(cmd.LogData)
	logdata.Grow(1)
	if mode == MODE_LINE {
		logdata.WriteByte('\n')
	} else {
		logdata.WriteByte(0)
	}
	logwriter.Write(logdata.Bytes())

	return nil
}

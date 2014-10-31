package gnlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/rare/gnet"
	log "github.com/cihub/seelog"
)

func validateDirAndFilename(catalog string, filename string) (string, string, error) {
	if len(catalog) > MAX_DIR_NAME_LEN {
		log.Warnf("dir name(%s) too long", catalog)
		return "", "", errors.New("dir name too long")
	}
	if len(filename) > MAX_FILE_NAME_LEN {
		log.Warnf("file name(%s) too long", filename)
		return "", "", errors.New("file name too long")
	}
	catalog = strings.Replace(catalog, strconv.QuoteRune(filepath.Separator), "_", -1)
	filename = strings.Replace(filename, strconv.QuoteRune(filepath.Separator), "_", -1)
	return catalog, filename, nil
}

func HandleStart(req *gnet.Request, resp *gnet.Response) error {
	log.Tracef("start handle StartCmd")

	buf, err := req.Body()
	if err != nil {
		log.Warnf("read start cmd body error")
		return err
	}

	cmd := new(StartCmd)
	if err := json.Unmarshal(buf, cmd); err != nil {
		log.Warnf("parse start cmd body error: (%v)", err)
		return err
	}

	var status int32 = STATUS_OK
	if err := Auth.Check(cmd.Auth); err != nil {
		log.Warnf("check auth error, auth string(%s)", cmd.Auth)
		status = STATUS_AUTH_ERR
		resp.SetCloseAfterSending()
	}

	var srbuf []byte
	sr := new(StartResp)
	sr.Status = status
	srbuf, err = json.Marshal(sr)
	if err != nil {
		log.Warnf("encoding StartResp error: (%v)", err)
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
		log.Warnf("recv LogCmd before channel authed")
		return errors.New("channel not authed")
	}

	buf, err := req.Body()
	if err != nil {
		log.Warnf("read LogCmd body error")
		return err
	}

	cmd := new(LogCmd)
	if err := json.Unmarshal(buf, cmd); err != nil {
		log.Warnf("parse LogCmd error: (%v)", err)
		return err
	}

	catalog, filename, err := validateDirAndFilename(cmd.Catalog, cmd.Filename)
	if err != nil {
		log.Warnf("validate dir(%s) and file(%s) name error", cmd.Catalog, cmd.Filename)
		return err
	}
	if len(cmd.LogData) == 0 {
		log.Warn("empty logdata, skip")
		return nil
	}

	logwriter := NewFileLogWriter()
	err = logwriter.Init(catalog, filename)
	if err != nil {
		log.Warnf("init log writer error: (%v)", err)
		return err
	}
	var mode uint8 = MODE_LINE
	if m, ok := req.Client().Storage().Get("mode"); ok {
		mode = m.(uint8)
	}
	logdata := bytes.NewBufferString(cmd.LogData)
	logdata.Grow(1)
	if mode == MODE_LINE {
		if cmd.LogData[len(cmd.LogData) - 1] != '\n' {
			logdata.WriteByte('\n')
		}
	} else {
		logdata.WriteByte(0)
	}
	logwriter.Write(logdata.Bytes())

	return nil
}

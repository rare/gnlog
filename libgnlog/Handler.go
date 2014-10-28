package gnlog

import (
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
	if cmd.Deserialize(buf) != nil {
		return err
	}

	if err := Auth.Check(cmd.Auth); err != nil {
		return err
	}

	catalog, filename, err := validateDirAndFilename(cmd.Catalog, cmd.Filename)
	if err != nil {
		return err
	}

	req.Client().Storage().Set("catalog", catalog)
	req.Client().Storage().Set("filename", filename)
	req.Client().Storage().Set("mode", cmd.Mode)

	logwriter := NewFileLogWriter()
	err = logwriter.Init(catalog, filename)
	if err != nil {
		return err
	}
	req.Client().Storage().Set("logwriter", logwriter)

	req.Client().Storage().Set("started", true)

	return nil
}

func HandleLog(req *gnet.Request, resp *gnet.Response) error {
	if _, ok := req.Client().Storage().Get("started"); !ok {
		return errors.New("channel not started")
	}

	buf, err := req.Body()
	if err != nil {
		return err
	}

	val, _ := req.Client().Storage().Get("logwriter")
	logwriter, _  := val.(LogWriter)
	logwriter.Write(buf)

	return nil
}

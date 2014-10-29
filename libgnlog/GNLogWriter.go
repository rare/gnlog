package gnlog

import (
	"path/filepath"
)

var (
	logwriterroutines	= make(map[string]*GNLogWriterRoutine)
)

type GNLogWriterRoutine struct {
	inchan		chan []byte

}

func NewGNLogWriterRoutine() *GNLogWriterRoutine {
	return &GNLogWriterRoutine {
		inchan:		make(chan []byte, Conf.LogChanBufSize),
	}
}

func (this *GNLogWriterRoutine) Init() error {

	return nil
}

func (this *GNLogWriterRoutine) Send(buf []byte) {
	this.inchan <- buf
}

type GNLogWriter struct {
	catalog		string
	filename	string
	routine		*GNLogWriterRoutine
}

func NewGNLogWriter() *GNLogWriter {
	return &GNLogWriter{
		catalog:	"",
		filename:	"",
		routine:	nil,
	}
}

func startGNLogWriter(catalog string, filename string) (*GNLogWriterRoutine, error) {
	return nil, nil
}

func (this *GNLogWriter) Init(catalog string, filename string) error {
	makeSureDirExists(filepath.Join(Conf.DataDir, this.catalog))
	makeSureFileExists(filepath.Join(Conf.DataDir, this.catalog, this.filename))

	routine, err := startGNLogWriter(this.catalog, this.filename)
	if err != nil {
		return err
	}
	this.routine = routine

	return nil
}

func (this *GNLogWriter) Write(buf []byte) error {
	this.routine.Send(buf)
	return nil
}

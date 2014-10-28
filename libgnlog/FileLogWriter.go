package gnlog

import (
	"path/filepath"
)

var (
	logwriterroutines	= make(map[string]*FileLogWriterRoutine)
)

type FileLogWriterRoutine struct {
	inchan		chan []byte

}

func NewFileLogWriterRoutine() *FileLogWriterRoutine {
	return &FileLogWriterRoutine {
		inchan:		make(chan []byte, Conf.LogChanBufSize),
	}
}

func (this *FileLogWriterRoutine) Init() error {

	return nil
}

func (this *FileLogWriterRoutine) Send(buf []byte) {
	this.inchan <- buf
}

type FileLogWriter struct {
	catalog		string
	filename	string
	routine		*FileLogWriterRoutine
}

func NewFileLogWriter() *FileLogWriter {
	return &FileLogWriter{
		catalog:	"",
		filename:	"",
		routine:	nil,
	}
}

func makeSureDirExists(path string) error {
	return nil
}

func makeSureFileExists(path string) error {
	return nil
}

func startFileLogWriter(catalog string, filename string) (*FileLogWriterRoutine, error) {
	return nil, nil
}

func (this *FileLogWriter) Init(catalog string, filename string) error {
	makeSureDirExists(filepath.Join(Conf.DataDir, this.catalog))
	makeSureFileExists(filepath.Join(Conf.DataDir, this.catalog, this.filename))

	routine, err := startFileLogWriter(this.catalog, this.filename)
	if err != nil {
		return err
	}
	this.routine = routine

	return nil
}

func (this *FileLogWriter) Write(buf []byte) error {
	this.routine.Send(buf)
	return nil
}

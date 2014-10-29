package gnlog

import (
	"os"
	"path/filepath"
	"sync"
	"time"
	"github.com/rare/gnet/gnutil"
)

var (
	fileMakeMutex		sync.Mutex
	fileLWMutex			sync.RWMutex
	fileLogWriters		= make(map[string]*FileLogWriterRoutine)
)

type FileLogWriterRoutine struct {
	inchan		chan []byte

}

func NewFileLogWriterRoutine() *FileLogWriterRoutine {
	return &FileLogWriterRoutine {
		inchan:		make(chan []byte, Conf.LogChanBufSize),
	}
}

func (this *FileLogWriterRoutine) Init(path string) error {
	go func() {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			return
		}

		for {
			buf := <-this.inchan
			_, err := f.Write(buf)
			if err != nil {
				break
			}

			fi, _ := os.Stat(path)
			if fi.Size() > Conf.MaxLogFileSize {
				f.Close()
				os.Rename(path, path + time.Now().Format("01-02-2006_03:04:55"))
				f, err = os.Create(path)
				if err != nil {
					//TODO
					break
				}
			}
		}

		f.Close()
	}()

	return nil
}

func (this *FileLogWriterRoutine) Send(buf []byte) {
	this.inchan<- buf
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
	exists := gnutil.DirExists(path)
	if !exists {
		os.Remove(path)

		fileMakeMutex.Lock()
		defer fileMakeMutex.Unlock()

		exists := gnutil.DirExists(path)
		if !exists {
			return os.MkdirAll(path, 0755)
		}
	}
	return nil
}

func makeSureFileExists(path string) error {
	exists := gnutil.FileExists(path)
	if !exists {
		os.Remove(path)

		fileMakeMutex.Lock()
		defer fileMakeMutex.Unlock()

		exists := gnutil.FileExists(path)
		if !exists {
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			f.Close()
		}
	}
	return nil
}

func startFileLogWriter(catalog string, filename string) (*FileLogWriterRoutine, error) {
	fileLWMutex.Lock()
	defer fileLWMutex.Unlock()

	flwr, ok := fileLogWriters[filepath.Join(catalog, filename)]
	if ok {
		return flwr, nil
	}

	flwr = NewFileLogWriterRoutine()
	if err := flwr.Init(filepath.Join(Conf.DataDir, catalog, filename)); err != nil {
		return nil, err
	}

	return flwr, nil
}

func (this *FileLogWriter) Init(catalog string, filename string) error {
	this.catalog = catalog
	this.filename = filename

	if err := makeSureDirExists(filepath.Join(Conf.DataDir, this.catalog)); err != nil {
		return err
	}
	if err := makeSureFileExists(filepath.Join(Conf.DataDir, this.catalog, this.filename)); err != nil {
		return err
	}

	fileLWMutex.RLock()
	flwr, ok := fileLogWriters[filepath.Join(this.catalog, this.filename)]
	fileLWMutex.RUnlock()

	if !ok {
		var err error
		this.routine, err = startFileLogWriter(this.catalog, this.filename)
		if err != nil {
			return err
		}
	} else {
		this.routine = flwr
	}

	return nil
}

func (this *FileLogWriter) Write(buf []byte) error {
	this.routine.Send(buf)
	return nil
}

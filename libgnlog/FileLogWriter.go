package gnlog

import (
	"fmt"		//debug

	"os"
	"path/filepath"
	"sync"
	"time"
	"github.com/rare/gnet/gnutil"
)

var (
	once				sync.Once
	fileMakeMutex		sync.Mutex
	fileLWMutex			sync.RWMutex
	fileLogWriters		= make(map[string]*FileLogWriterRoutine)
)

func runSplitFileMonitor() {
	ticker := time.NewTicker(time.Second)
	lasttime := time.Now()

	for {
		now := <-ticker.C
		fileLWMutex.RLock()

		for _, flw := range fileLogWriters {
			//split by time
			if Conf.Log.SplitPolicy.ByTime.Enable {
				rule := Conf.Log.SplitPolicy.ByTime.Rule
				if ((rule == "byday" && now.Day() != lasttime.Day()) || (rule == "byhour" && now.Hour() != lasttime.Hour())) {
					flw.spchan<- now
				}
			}

			//split by size
			if Conf.Log.SplitPolicy.BySize.Enable {
				fi, err := os.Stat(flw.path)
				if err != nil {
					//TODO
					continue
				}

				if fi.Size() > Conf.Log.SplitPolicy.BySize.MaxLogFileSize {
					flw.spchan<- now
				}
			}
		}

		fileLWMutex.RUnlock()
		lasttime = now
	}
}

type FileLogWriterRoutine struct {
	path		string
	inchan		chan []byte
	spchan		chan time.Time
}

func NewFileLogWriterRoutine() *FileLogWriterRoutine {
	return &FileLogWriterRoutine {
		path:		"",
		inchan:		make(chan []byte, Conf.Log.BufSize),
		spchan:		make(chan time.Time),
	}
}

func (this *FileLogWriterRoutine) Run() {
	//TODO
	//trace
	fmt.Println("start file log writer routine")

	f, err := os.OpenFile(this.path, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		//TODO
		//debug
		fmt.Println("open file(" + this.path + "for writing error")
		return
	}

	for {
		select {
			case buf := <-this.inchan:
				_, err := f.Write(buf)
				if err != nil {
					//TODO
					break
				}
			case t := <-this.spchan:
				f.Close()
				newpath := fmt.Sprintf("%s.%d%d%d%d%d%d", this.path, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
				os.Rename(this.path, newpath)
				f, err = os.Create(this.path)
				if err != nil {
					//TODO
					break
				}
		}
	}

	f.Close()
}

func (this *FileLogWriterRoutine) Init(path string) {
	this.path = path
	go this.Run()
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
	flwr.Init(filepath.Join(Conf.Log.Dir, catalog, filename))
	fileLogWriters[filepath.Join(catalog, filename)] = flwr

	return flwr, nil
}

func (this *FileLogWriter) Init(catalog string, filename string) error {
	once.Do(func(){go runSplitFileMonitor()})

	this.catalog = catalog
	this.filename = filename

	if err := makeSureDirExists(filepath.Join(Conf.Log.Dir, this.catalog)); err != nil {
		//TODO
		return err
	}
	if err := makeSureFileExists(filepath.Join(Conf.Log.Dir, this.catalog, this.filename)); err != nil {
		//TODO
		return err
	}

	fileLWMutex.RLock()
	flwr, ok := fileLogWriters[filepath.Join(this.catalog, this.filename)]
	fileLWMutex.RUnlock()

	if !ok {
		var err error
		this.routine, err = startFileLogWriter(this.catalog, this.filename)
		if err != nil {
			//TODO
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

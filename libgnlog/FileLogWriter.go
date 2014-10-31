package gnlog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"github.com/rare/gnet/gnutil"
	log "github.com/cihub/seelog"
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
					log.Warnf("split by size, stat file(%s) error: (%v)", flw.path, err)
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
	log.Tracef("start file log writer routine for file(%s)", this.path)

	f, err := os.OpenFile(this.path, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Warnf("open file(%s) for writing error: (%v)", this.path, err)
		return
	}
	wr := bufio.NewWriterSize(f, int(Conf.Log.BufSize * 4096))

	for {
		select {
			case buf := <-this.inchan:
				_, err := wr.Write(buf)
				if err != nil {
					log.Warnf("write data to file(%s) error: (%v)", this.path, err)
					break
				}
			case t := <-this.spchan:
				wr.Flush()
				f.Close()
				newpath := fmt.Sprintf("%s.%d%d%d%d%d%d", this.path, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
				os.Rename(this.path, newpath)
				f, err = os.Create(this.path)
				if err != nil {
					log.Warnf("create file(%s) after split error: (%v)", this.path, err)
					return
				}
				wr = bufio.NewWriterSize(f, int(Conf.Log.BufSize * 4096))
		}
	}

	wr.Flush()
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
		log.Warnf("init file log writer, create dir(%s) error: (%v)", this.catalog, err)
		return err
	}
	if err := makeSureFileExists(filepath.Join(Conf.Log.Dir, this.catalog, this.filename)); err != nil {
		log.Warnf("init file log writer, create file(%s) error: (%v)", this.filename, err)
		return err
	}

	fileLWMutex.RLock()
	flwr, ok := fileLogWriters[filepath.Join(this.catalog, this.filename)]
	fileLWMutex.RUnlock()

	if !ok {
		var err error
		this.routine, err = startFileLogWriter(this.catalog, this.filename)
		if err != nil {
			log.Warnf("start file log writer error: (%v)", err)
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

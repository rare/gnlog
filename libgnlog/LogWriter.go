package gnlog

import (
	"bufio"
	"bytes"
)

type LogWriter struct {
	filename	string
	wr			*bufio.Writer			
}

func NewLogWriter() *LogWriter {
	return &LogWriter{
	}
}

func (this *LogWriter) Write(buf *bytes.Buffer) error {
	this.wr.Write(buf.Bytes())
	return nil
}

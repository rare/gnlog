package gnlog

type LogWriter interface {
	Write(buf []byte) error
}

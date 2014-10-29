package gnlogcli

import (
	"bytes"
	"encoding/json"
	"net"
	"time"
	"github.com/rare/gnet/gnproto"
	"github.com/rare/gnlog/libgnlog"
)

const (
	HEART_BEAT_INTERVAL			= 5
	MAX_OUT_CHAN_BUF_SIZE		= 100
	WRITE_TIMEOUT				= 3
)

type GNLogClient struct {
	exit		chan bool
	conn		net.Conn
	nextseq		uint32
	outchan		chan []byte
	auth		string
	catalog		string
	filename	string
}

func NewGNLogClient() *GNLogClient {
	return &GNLogClient {
		exit:		make(chan bool),
		conn:		nil,
		nextseq:	0,
		outchan:	make(chan *bytes.Buffer, MAX_OUT_CHAN_BUF_SIZE),
	}
}

func (this *GNLogClient) handleInput() {
}

func (this *GNLogClient) handleOutput() {
	for {
		select {
			case buf := <-this.outchan:
				this.conn.SetWriteDeadline(time.Now().Add(time.Duration(WRITE_TIMEOUT) * time.Second))
				if _, err := this.conn.Write(buf.Bytes()); err != nil {
					close(this.exit)
				}
			case <-this.exit:
				return
		}
	}
}

func (this *GNLogClient) sendHeartbeat() {
	ticker := time.NewTicker(time.Duration(HEART_BEAT_INTERVAL) * time.Second)

	for {
		select {
		case <-ticker.C:
			var header gnproto.Header = gnproto.Header {
				Cmd:	gnproto.CMD_HEART_BEAT,
				Len:	0,
				Seq:	0,
				Ver:	0,
			}
			buf, _ := header.Serialize()
			this.outchan<- bytes.NewBuffer(buf)
		case <-this.exit:
			return
		}
	}
}

func (this *GNLogClient) Init(addr, auth, catalog, filename string) error {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return err
	}
	this.conn = conn
	this.auth = auth
	this.catalog = catalog
	this.filename = filename

	go this.handleOutput()

	cmd := new(gnlog.StartCmd)
	cmd.Auth = this.auth
	cmd.Mode = gnlog.MODE_LINE
	body, err := json.Marshal(cmd)

	hdr := new(gnproto.Header)
	hdr.Cmd = gnlog.CMD_START
	hdr.Ver = 0
	hdr.Seq = this.nextseq; this.nextseq++
	hdr.Len = (uint32)(len(body))
	var header []byte
	header, err = hdr.Serialize()
	if err != nil {
		return err
	}

	buf := bytes.Newbuffer(header)
	buf.Grow(len(body))
	buf.Write(body)
	this.outchan<- buf

	//recv ack
	//TODO

	go this.sendHeartbeat()

	return nil
}

func (this *GNLogClient) Log(msg string) error {
	cmd := new(gnlog.LogCmd)
	cmd.Catalog = this.catalog
	cmd.Filename = this.filename
	cmd.LogData = msg
	body, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	hdr := new(gnproto.Header)
	hdr.Cmd = gnlog.CMD_LOG
	hdr.Ver = 0
	hdr.Seq = this.nextseq; this.nextseq++
	hdr.Len = (uint32)(len(body))
	var header []byte
	header, err = hdr.Serialize()
	if err != nil {
		return err
	}

	buf := bytes.Newbuffer(header)
	buf.Grow(len(body))
	buf.Write(body)
	this.outchan<- buf

	return nil
}

package gnlogcli

import (
	"fmt"		//debug

	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"time"
	"github.com/rare/gnet/gnproto"
	"github.com/rare/gnlog/libgnlog"
)

const (
	HEART_BEAT_INTERVAL			= 5
	MAX_OUT_CHAN_BUF_SIZE		= 100
	WRITE_TIMEOUT				= 30
	READ_TIMEOUT				= 30
)

type GNLogClient struct {
	exit		chan bool
	conn		net.Conn
	nextseq		uint32
	outchan		chan *bytes.Buffer
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

func (this *GNLogClient) handleOutput() {
	for {
		select {
			case buf := <-this.outchan:
				//debug
				this.conn.SetWriteDeadline(time.Now().Add(time.Duration(WRITE_TIMEOUT) * time.Second))
				if _, err := this.conn.Write(buf.Bytes()); err != nil {
					//debug
					fmt.Printf("write error(%s), close exit chan\n", err.Error())
					close(this.exit)
					return
				}
			case <-this.exit:
				//debug
				fmt.Println("exit output routine")
				return
		}
	}
}

func (this *GNLogClient) sendHeartbeat() {
	ticker := time.NewTicker(time.Duration(HEART_BEAT_INTERVAL) * time.Second)

	for {
		select {
		case <-ticker.C:
			//debug
			fmt.Println("ticker ticked")
			var header gnproto.Header = gnproto.Header {
				Cmd:	gnproto.CMD_HEART_BEAT,
				Len:	0,
				Seq:	0,
				Ver:	0,
			}
			buf, _ := header.Serialize()
			//debug
			fmt.Println("send heartbeat to outchan")
			this.outchan<- bytes.NewBuffer(buf)
			//debug
			fmt.Println("send heart beat")
		case <-this.exit:
			//debug
			fmt.Println("exit heartbeat routine")
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

	buf := bytes.NewBuffer(header)
	buf.Grow(len(body))
	buf.Write(body)
	this.outchan<- buf

	this.conn.SetReadDeadline(time.Now().Add(time.Duration(READ_TIMEOUT) * time.Second))
	var headbuf []byte = make([]byte, gnproto.HEADER_SIZE)
	n, err := io.ReadFull(this.conn, headbuf)
	if n != len(headbuf) {
		return err
	}
	if err := hdr.Deserialize(headbuf); err != nil {
		return err
	}
	if hdr.Len <= 0 {
		return errors.New("invalid response for StartCmd")
	}
	var bodybuf []byte = make([]byte, hdr.Len)
	n, err = io.ReadFull(this.conn, bodybuf)
	if n != len(bodybuf) {
		return err
	}
	resp := new(gnlog.StartResp)
	err = json.Unmarshal(bodybuf, resp)
	if err != nil {
		return err
	}
	if resp.Status != gnlog.STATUS_OK {
		return errors.New("start error")
	}

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

	buf := bytes.NewBuffer(header)
	buf.Grow(len(body))
	buf.Write(body)
	this.outchan<- buf

	return nil
}

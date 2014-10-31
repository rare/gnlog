package gnlogcli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
	"github.com/rare/gnet/gnproto"
	"github.com/rare/gnlog/libgnlog"
	log "github.com/cihub/seelog"
)

const (
	HEART_BEAT_INTERVAL			= 5
	MAX_OUT_CHAN_BUF_SIZE		= 100
	WRITE_TIMEOUT				= 30
	READ_TIMEOUT				= 30
)

type Config struct {
	Addr		string		`json:"addr"`
	Auth		string		`json:"auth"`
	Catalog		string		`json:"catalog"`
	Filename	string		`json:"filename"`
	Minlevel	uint8		`json:"minlevel"`
	Format		string		`json:"format"`
}

var (
	Conf Config
)



type GNLogClient struct {
	exit		chan bool
	conn		net.Conn
	nextseq		uint32
	logchan		chan string
	hbchan		chan bool
	auth		string
	catalog		string
	filename	string
}

func NewGNLogClient() *GNLogClient {
	return &GNLogClient {
		exit:		make(chan bool),
		conn:		nil,
		nextseq:	0,
		logchan:	make(chan string, MAX_OUT_CHAN_BUF_SIZE),
		hbchan:		make(chan bool),
		auth:		"",
		catalog:	"",
		filename:	"",
	}
}

func (this *GNLogClient) sendLog(msg string) error {
	cmd := new(gnlog.LogCmd)
	cmd.Catalog = this.catalog
	cmd.Filename = this.filename
	cmd.LogData = msg
	body, err := json.Marshal(cmd)
	if err != nil {
		fmt.Printf("serialize log cmd body error: (%v)\n", err)
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
		fmt.Printf("serialize log cmd header error: (%v)\n", err)
		return err
	}

	buf := bytes.NewBuffer(header)
	buf.Grow(len(body))
	buf.Write(body)

	this.conn.SetWriteDeadline(time.Now().Add(time.Duration(WRITE_TIMEOUT) * time.Second))
	if _, err := this.conn.Write(buf.Bytes()); err != nil {
		fmt.Printf("conn(%p) write error(%v), close exit chan\n", this.conn, err)
		close(this.exit)
		return errors.New("write error")
	}

	return nil
}

func (this *GNLogClient) sendHeartbeat() error {
	var header gnproto.Header = gnproto.Header {
		Cmd:	gnproto.CMD_HEART_BEAT,
		Len:	0,
		Seq:	0,
		Ver:	0,
	}
	buf, _ := header.Serialize()

	this.conn.SetWriteDeadline(time.Now().Add(time.Duration(WRITE_TIMEOUT) * time.Second))
	if _, err := this.conn.Write(buf); err != nil {
		fmt.Printf("conn(%p) write error(%v), close exit chan\n", this.conn, err)
		close(this.exit)
		return errors.New("write error")
	}

	return nil
}

func (this *GNLogClient) handleOutput() {
	for {
		select {
			case msg := <-this.logchan:
				this.sendLog(msg)
			case <-this.hbchan:
				this.sendHeartbeat()
			case <-this.exit:
				fmt.Println("exit output routine")
				return
		}
	}
}

func (this *GNLogClient) handleHeartbeat() {
	ticker := time.NewTicker(time.Duration(HEART_BEAT_INTERVAL) * time.Second)

	for {
		select {
		case <-ticker.C:
			this.hbchan<- true
		case <-this.exit:
			fmt.Println("exit heartbeat routine")
			return
		}
	}
}

func (this *GNLogClient) Init(addr, auth, catalog, filename string) error {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		fmt.Printf("dial tcp error: (%v)\n", err)
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
	if err != nil {
		fmt.Printf("serialize start cmd body error: (%v)\n", err)
		return err
	}

	hdr := new(gnproto.Header)
	hdr.Cmd = gnlog.CMD_START
	hdr.Ver = 0
	hdr.Seq = this.nextseq; this.nextseq++
	hdr.Len = (uint32)(len(body))
	var header []byte
	header, err = hdr.Serialize()
	if err != nil {
		fmt.Printf("serialize start cmd header error: (%v)\n", err)
		return err
	}

	buf := bytes.NewBuffer(header)
	buf.Grow(len(body))
	buf.Write(body)
	this.conn.SetWriteDeadline(time.Now().Add(time.Duration(WRITE_TIMEOUT) * time.Second))
	if _, err := this.conn.Write(buf.Bytes()); err != nil {
		fmt.Printf("conn(%p) write error(%v), close exit chan\n", this.conn, err)
		close(this.exit)
		return errors.New("write error")
	}

	this.conn.SetReadDeadline(time.Now().Add(time.Duration(READ_TIMEOUT) * time.Second))
	var headbuf []byte = make([]byte, gnproto.HEADER_SIZE)
	n, err := io.ReadFull(this.conn, headbuf)
	if n != len(headbuf) {
		fmt.Printf("read start resp error: (%v)\n", err)
		return err
	}
	if err := hdr.Deserialize(headbuf); err != nil {
		fmt.Printf("parse start resp error: (%v)\n", err)
		return err
	}
	if hdr.Len <= 0 {
		fmt.Printf("invalid data length(%d) of start\n", hdr.Len)
		return errors.New("invalid response for StartCmd")
	}
	var bodybuf []byte = make([]byte, hdr.Len)
	n, err = io.ReadFull(this.conn, bodybuf)
	if n != len(bodybuf) {
		fmt.Printf("read start resp body error: (%v)\n", err)
		return err
	}
	resp := new(gnlog.StartResp)
	err = json.Unmarshal(bodybuf, resp)
	if err != nil {
		fmt.Printf("parse start resp body error: (%v)\n", err)
		return err
	}
	if resp.Status != gnlog.STATUS_OK {
		fmt.Printf("start resp status(%d) not ok\n", resp.Status)
		return errors.New("start error")
	}

	go this.handleHeartbeat()

	return nil
}

func (this *GNLogClient) Log(msg string) error {
	this.logchan<- msg
	return nil
}

func initLogWriter(cli *GNLogClient, lvl uint8, format string) error {
	logger, err := log.LoggerFromWriterWithMinLevelAndFormat(NewGNLogWriter(cli), log.LogLevel(lvl), format)
	if err != nil {
		fmt.Printf("init seelog logger error: (%v)\n", err)
		return err
	}
	log.ReplaceLogger(logger)
	return nil
}

func Init(conf *Config) error {
	Conf = *conf

	cli := NewGNLogClient()
	if err := cli.Init(Conf.Addr, Conf.Auth, Conf.Catalog, Conf.Filename); err != nil {
		fmt.Printf("init gnlog client error: (%v)\n", err)
		return err
	}

	return initLogWriter(cli, Conf.Minlevel, Conf.Format)
}


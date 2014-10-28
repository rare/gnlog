package gnlog

import (
	"bytes"
	"encoding/binary"
)

const (
	CMD_START			=	uint16(1)	//StartCmd, json format
	CMD_LOG				=	uint16(2)	//raw data
)

const (
	MODE_LINE			=	uint8(1)	//each log end by \n
	MODE_STRING			=	uint8(2)	//each log end by \0
	MODE_BINARY			=	uint8(3)	//each log is a binary struct
)

type StartCmd struct {
	Auth		string					`json:"auth"`
	Catalog		string					`json:"catalog"`
	Filename	string					`json:"filename"`
	Mode		uint8					`json:"mode"`
}

func (this *StartCmd) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, *this); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *StartCmd) Deserialize(b []byte) error {
	buf := bytes.NewReader(b)
	if err := binary.Read(buf, binary.BigEndian, this); err != nil {
		return err
	}
	return nil
}

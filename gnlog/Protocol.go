package gnlog

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
	FileName	string					`json:"filename"`
	Mode		uint8					`json:"mode"`
}


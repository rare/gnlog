package gnlog

const (
	CMD_START			=	uint16(1)	//StartCmd, json format
	CMD_LOG				=	uint16(2)	//raw data
)

const (
	MODE_LINE			=	uint8(1)	//each log end by \n
	MODE_STRING			=	uint8(2)	//each log end by \0
)

const (
	STATUS_OK			=	int32(0)
	STATUS_AUTH_ERR		=	int32(-1)
)

type StartCmd struct {
	Auth		string					`json:"auth"`
	Mode		uint8					`json:"mode"`
}

type StartResp struct {
	Status		int32					`json:"status"`
}

type LogCmd struct {
	Catalog		string					`json:"catalog"`
	Filename	string					`json:"filename"`
	LogData		string					`json:"logdata"`
}

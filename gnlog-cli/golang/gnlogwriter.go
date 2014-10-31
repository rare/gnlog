package gnlogcli

type GNLogWriter struct {
	client	*GNLogClient

}

func NewGNLogWriter(client *GNLogClient) *GNLogWriter {
	return &GNLogWriter {
		client:		client,
	}
}

func (this *GNLogWriter) Write(p []byte) (n int, err error) {
	return len(p), this.client.Log(string(p))
}

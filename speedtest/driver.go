package speedtest

import "github.com/angrylid/clashtester/tester"

type Speedtest struct {
	server *Server
}

func UseSpeedtest() *Speedtest {
	return &Speedtest{}
}

func (sp *Speedtest) Prepare() error {
	server, err := SelectNearestServer()
	if err != nil {
		return err
	}
	sp.server = server
	return nil
}

func (sp *Speedtest) GetLatency() error {
	_, _, err := sp.server.GetPingLatency()
	if err != nil {
		return err
	}
	return nil
}

func (sp *Speedtest) MakeDownload() error {
	panic("not implemented") // TODO: Implement
}

func (sp *Speedtest) MakeUpload() error {
	panic("not implemented") // TODO: Implement
}

func (sp *Speedtest) GetUserInfo() error {
	panic("not implemented") // TODO: Implement
}

func (sp *Speedtest) GetResult() tester.Result {
	panic("not implemented") // TODO: Implement
}

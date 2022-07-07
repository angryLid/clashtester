package tester

import (
	"time"
)

type Driver interface {
	GetLatency() error
	MakeDownload() error
	MakeUpload() error
	GetUserInfo() error
	GetResult() Result
	Prepare() error
}

func (mock *Mock) Prepare() error {
	//
	panic("not implemented") // TODO: Implement
}

type DriverFactory func() Driver

type Mock struct {
	latency       time.Duration
	downloadSpeed float64
	uploadSpeed   float64
	cfg           string
}

func UseMock() *Mock {
	return &Mock{}
}

func (m *Mock) GetLatency() error {
	m.latency = time.Millisecond * 136
	return nil
}
func (m *Mock) MakeDownload() error {
	m.downloadSpeed = 5.42
	return nil
}
func (d *Mock) MakeUpload() error {
	d.uploadSpeed = 3.79
	return nil
}
func (d *Mock) GetUserInfo() error {
	d.cfg = "A server in Hong Kong."
	return nil
}
func (d *Mock) GetResult() Result {
	return Result{
		d.latency, d.latency, d.downloadSpeed, d.uploadSpeed,
	}
}

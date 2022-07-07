package tester

import (
	"errors"
	"time"
)

const (
	DriverMock = iota
	DriverSpeedtest
	DriverCloudflare
)

type Result struct {
	TCPLatency    time.Duration
	HTTPLatency   time.Duration
	DownloadSpeed float64
	UploadSpeed   float64
}
type Tester struct {
	item   int
	driver Driver
}

func NewTester(driver Driver) *Tester {
	return &Tester{
		ItemLatency | ItemDownload,
		driver,
	}
}

func (t *Tester) SetItem(item int) error {
	if item < 1 || item > 15 {
		return errors.New("invalid arguments of SetItem")
	}
	t.item = item
	return nil
}

func (t *Tester) Do() (Result, error) {
	var driver = t.driver
	var itemHandlers = [...]ItemHandler{
		{ItemUserInfo, driver.GetUserInfo},
		{ItemLatency, driver.GetLatency},
		{ItemDownload, driver.MakeDownload},
		{ItemUpload, driver.MakeUpload},
	}
	driver.Prepare()
	for _, ih := range itemHandlers {
		if t.item&ih.Item != 0 {
			ih.Handler()
		}
	}
	return driver.GetResult(), nil
}

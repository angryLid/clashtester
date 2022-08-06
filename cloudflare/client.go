package cloudflare

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"
)

type HTTPClient struct {
	*http.Client
}

func NewHTTPClient(proxy string) *HTTPClient {
	p, _ := url.Parse(proxy)
	tr := &http.Transport{
		Proxy:           http.ProxyURL(p),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &HTTPClient{
		&http.Client{
			Transport: tr,
			Timeout:   time.Second * 15,
		},
	}
}

func (c *HTTPClient) Download(url string) (*Tracer, error) {
	tracer := &Tracer{}
	trace := &httptrace.ClientTrace{
		GetConn:              tracer.GetConn,
		GotConn:              tracer.GotConn,
		GotFirstResponseByte: tracer.GotFirstResponseByte,
	}

	traceCtx := httptrace.WithClientTrace(context.Background(), trace)
	req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	written, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return nil, err
	}
	tracer.GotResponseBody(written)
	return tracer, nil
}

func (c *HTTPClient) Upload(url string, size int64) (*Speed, error) {

	buf := bytes.NewBuffer(make([]byte, size))
	req, _ := http.NewRequest(http.MethodGet, url, buf)

	beforeReq := time.Now()
	res, err := c.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	_, err = io.Copy(ioutil.Discard, res.Body)

	if err != nil {
		return nil, err
	}
	duration := time.Since(beforeReq)

	return &Speed{size, duration}, nil
}

type Cloudflare struct {
	*HTTPClient
	measureID int
}

func New(proxy string) *Cloudflare {
	return &Cloudflare{
		HTTPClient: NewHTTPClient(proxy),
		measureID:  MeasureID(),
	}
}

func (cf *Cloudflare) Download(size int64) (*Tracer, error) {
	// url := fmt.Sprintf("https://speed.cloudflare.com/__down?measId=%dbytes=%d", cf.measureID, size)
	url := fmt.Sprintf("https://speed.cloudflare.com/__down?bytes=%d", size)
	tracer, err := cf.HTTPClient.Download(url)
	if err != nil {
		return tracer, err
	}
	tracer.written = size
	return tracer, err
}

func (cf *Cloudflare) Upload(size int64) (*Speed, error) {
	// url := fmt.Sprintf("https://speed.cloudflare.com/__up?measId=%d", cf.measureID)
	url := "https://speed.cloudflare.com/__up"
	return cf.HTTPClient.Upload(url, size)
}

type Meta struct {
	Hostname       string `json:"hostname"`
	ClientIP       string `json:"clientIp"`
	HTTPProtocol   string `json:"httpProtocol"`
	Asn            int    `json:"asn"`
	AsOrganization string `json:"asOrganization"`
	Colo           string `json:"colo"`
	Country        string `json:"country"`
	City           string `json:"city"`
	Region         string `json:"region"`
	PostalCode     string `json:"postalCode"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
}

func (m Meta) String() string {
	return fmt.Sprintf("[%s, %s, %s]", m.ClientIP, m.AsOrganization, m.Colo)
}

func (c *Cloudflare) GetMeta() (*Meta, error) {
	var MetaURL = "https://speed.cloudflare.com/meta"
	req, _ := http.NewRequest(http.MethodGet, MetaURL, nil)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	stream, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var meta Meta
	err = json.Unmarshal(stream, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func MeasureID() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(9000_0000_0000_0000) + 1000_0000_0000_0000
}

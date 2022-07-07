package speedtest

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	proxy, _ := url.Parse("http://localhost:7890")
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5, //超时时间
	}

	client := NewClient(httpClient)

	client.FetchUserInfo()
	servers, err := client.FetchServerList(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, server := range servers {
		server.GetPingLatency()
		log.Println(server.Name, server.Latency)
	}
}

func TestSelectClosestServer(t *testing.T) {
	c := Setup()
	serverlist, err := c.FetchServerList(context.Background())

	if err != nil {
		t.Fatal("Fetch Server List Failed.")
	}

	server := serverlist[0]
	channel, err := server.DownLoadTest(context.Background(), c.inner)
	if err != nil {
		log.Fatal(err)
	}

	<-channel
	t.Log(server.DLSpeed)
}

func TestDownload(t *testing.T) {
	c := Setup()
	panic(c)
}

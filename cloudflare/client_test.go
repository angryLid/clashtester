package cloudflare

import (
	"testing"

	"github.com/angrylid/clashtester/clash"
)

func TestGetMeta(t *testing.T) {
	client := New("http://localhost:7890")
	clash.SetMode(clash.GlobalMode)
	defer clash.SetMode(clash.RuleMode)

	meta, err := client.GetMeta()
	t.Log(meta, err)
}

func TestUpload(t *testing.T) {
	client := NewHTTPClient("http://localhost:7890")
	clash.SetMode(clash.GlobalMode)
	defer clash.SetMode(clash.RuleMode)

	speed, err := client.Upload("https://speed.cloudflare.com/__up", 100*1000)
	t.Log(speed, err)
}

func TestDownload(t *testing.T) {

}

// Normal Test: 20Ping, 10*100KB, 8*1MB, 6*10MB, 8*100KB, 6*1MB, 4*10MB
func TestBenchmark(t *testing.T) {

}

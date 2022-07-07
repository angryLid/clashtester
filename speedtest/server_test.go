package speedtest

import (
	"log"
	"strconv"
	"testing"
)

func TestSelectBestServer(t *testing.T) {
	server, err := SelectNearestServer()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(server)
}

func TestMakeDownload(t *testing.T) {
	server, err := SelectNearestServer()
	if err != nil {
		t.Fatal(err)
	}

	speed, err := server.MakeDownload()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(speed)
}

func TestServer_MakeUpload(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	server, err := SelectNearestServer()
	if err != nil {
		t.Fatal(err)
	}

	result, _ := server.MakeUpload()
	speed, _ := result.GetSpeed()
	t.Log(result.ContentLength, result.Spent, speed)
}

func TestByteSlice(t *testing.T) {
	slice := []byte{'s', 'i', 'z', 'e', '=', '1', '0', '0', '\n'}
	t.Log(strconv.ParseInt(string(slice[5:len(slice)-1]), 10, 64))
}

package speedtest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptrace"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	URL       string  `xml:"url,attr" json:"url"`
	Lat       string  `xml:"lat,attr" json:"lat"`
	Lon       string  `xml:"lon,attr" json:"lon"`
	Distance  float64 `json:"distance"`
	Name      string  `xml:"name,attr" json:"name"`
	Country   string  `xml:"country,attr" json:"country"`
	CC        string  `json:"cc"`
	Sponsor   string  `xml:"sponsor,attr" json:"sponsor"`
	ID        string  `xml:"id,attr" json:"id"`
	Preferred float64 `json:"preferred"`
	Host      string  `xml:"host,attr" json:"host"`

	// Optional. Server with these key may be lack of some utilities.
	ForcePingSelect float64 `json:"force_ping_select"`
	HttpsFunctional float64 `json:"https_functional"`

	// Deprecated
	URL2 string `xml:"url2,attr" json:"url_2"`
	// Deprecated
	Latency time.Duration `json:"latency"`
	// Deprecated
	DLSpeed float64 `json:"dl_speed"` // Mbps
	// Deprecated
	ULSpeed float64 `json:"ul_speed"` // Mbps
}

type ServerList []*Server

type Result struct {
	ContentLength int64
	Spent         time.Duration
}

func (r Result) GetSpeedMbps() (float64, error) {
	speed, err := r.GetSpeed()
	if err != nil {
		return 0, err
	}
	return speed * 8, nil
}

func (r Result) GetSpeed() (float64, error) {

	if r.Spent == 0 || r.ContentLength == 0 {
		return 0, fmt.Errorf("invalid result")
	}
	downloaded := float64(r.ContentLength) / (1024 * 1024)
	spent := float64(r.Spent) / float64(time.Second)

	return downloaded / spent, nil
}

func (s *Server) String() string {
	return fmt.Sprintf("Distance: %f <%s \"%s\"> host: %s \n", s.Distance, s.CC, s.Sponsor, s.Host)
}
func FetchServerList() (ServerList, error) {
	req, err := http.NewRequest(http.MethodGet, speedTestServersUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serverList ServerList

	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&serverList); err != nil {
		return serverList, err
	}
	if len(serverList) == 0 {
		return nil, errors.New("failed to fetch servers")
	}

	return serverList, nil
}

func SelectNearestServer() (server *Server, err error) {
	serverList, err := FetchServerList()
	if err != nil {
		return nil, err
	}
	sort.Slice(serverList, func(i, j int) bool {
		return serverList[i].Distance < serverList[j].Distance
	})
	return serverList[0], nil
}

func (s *Server) GetPingLatency() (time.Duration, time.Duration, error) {
	pingURL := strings.Split(s.URL, "/upload.php")[0] + "/latency.txt"

	tracer := new(Tracer)

	trace := &httptrace.ClientTrace{
		GetConn:              tracer.GetConn,
		GotConn:              tracer.GotConn,
		GotFirstResponseByte: tracer.GotFirstResponseByte,
	}
	req, _ := http.NewRequest(http.MethodGet, pingURL, nil)

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := httpClient.Transport.RoundTrip(req)

	if err != nil {
		return 0, 0, err
	}
	resp.Body.Close()
	// trans to  ms
	s.Latency = tracer.latency
	return tracer.totalLatency, tracer.latency, nil

}

func (s *Server) MakeDownload() (Result, error) {

	const size = 50 * 1000000
	url := fmt.Sprintf("https://%s/download?nocache=%s&guid=%s&size=%d", s.Host, uuid.New(), uuid.New(), size)

	tracer := new(Tracer)

	trace := &httptrace.ClientTrace{
		GetConn:              tracer.GetConn,
		GotConn:              tracer.GotConn,
		GotFirstResponseByte: tracer.GotFirstResponseByte,
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	start := time.Now()
	resp, err := httpClient.Transport.RoundTrip(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return Result{}, err
	}
	end := time.Since(start)

	return Result{
		ContentLength: resp.ContentLength,
		Spent:         end,
	}, nil
}

func (s *Server) MakeUpload() (Result, error) {
	url := fmt.Sprintf("https://%s/upload?nocache=%s&guid=%s", s.Host, uuid.New(), uuid.New())

	buf := bytes.NewBuffer(make([]byte, 75*1000000))
	req, _ := http.NewRequest(http.MethodPost, url, buf)

	req.Header.Set("Content-Type", "application/octet-stream")

	start := time.Now()
	resp, err := httpClient.Do(req)
	if err != nil {
		return Result{}, err
	}
	stream, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Result{}, err
	}
	byteCount, _ := strconv.ParseInt(string(stream[5:len(stream)-1]), 10, 64)
	defer resp.Body.Close()

	end := time.Since(start)
	return Result{
		ContentLength: byteCount,
		Spent:         end,
	}, nil
}

func (s *Server) DownLoadTest(ctx context.Context, c *http.Client) (chan DeprecatedResult, error) {
	// TODO config this
	size := dlSizes[2] // 750*750.jpg ~= 500k one request
	threadCount := 1
	resChan := make(chan DeprecatedResult, 1)

	// base download url
	// serverhost/random/750x750.jpg
	dlURL := strings.Split(s.URL, "/upload.php")[0] + "/random" + strconv.Itoa(size) + "x" + strconv.Itoa(size) + ".jpg"

	eg, ctx := errgroup.WithContext(ctx)
	respBytes := atomic.NewInt64(0)
	sTime := time.Now()
	for i := 0; i < threadCount; i++ {
		eg.Go(func() error {
			for i := 0; i < 10; i++ {
				s, err := downloadRequest(ctx, c, dlURL)
				if err == nil {
					respBytes.Add(s)
					resChan <- DeprecatedResult{CurrentSpeed: calcMbpsSpeed(respBytes.Load(), sTime), CurrentBytes: respBytes.Load()}
				} else {
					return err
				}
			}
			return nil
		})
	}

	// start speed test thread
	go func() {
		if err := eg.Wait(); err != nil {
			// TODO add err ch
			println(err.Error())
		}
		close(resChan)
		s.DLSpeed = calcMbpsSpeed(respBytes.Load(), sTime)
	}()

	return resChan, nil
}

func downloadRequest(ctx context.Context, c *http.Client, dlURL string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dlURL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	return resp.ContentLength, err
}

// Len finds length of servers. For sorting servers.
func (svrs ServerList) Len() int {
	return len(svrs)
}

// Swap swaps i-th and j-th. For sorting servers.
func (svrs ServerList) Swap(i, j int) {
	svrs[i], svrs[j] = svrs[j], svrs[i]
}

// Swap swaps i-th and j-th. For sorting servers.
func (svrs ServerList) Less(i, j int) bool {
	return svrs[i].Distance < svrs[j].Distance
}

func Distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	radius := 6378.137

	a1 := lat1 * math.Pi / 180.0
	b1 := lon1 * math.Pi / 180.0
	a2 := lat2 * math.Pi / 180.0
	b2 := lon2 * math.Pi / 180.0

	x := math.Sin(a1)*math.Sin(a2) + math.Cos(a1)*math.Cos(a2)*math.Cos(b2-b1)
	return radius * math.Acos(x)
}

type DeprecatedResult struct {
	CurrentSpeed float64
	CurrentBytes int64
}

func calcMbpsSpeed(bytes int64, startTime time.Time) float64 {
	fTime := time.Now()
	// MBps(MB per second)
	MBps := float64(bytes) / 1000 / 1000 / fTime.Sub(startTime).Seconds()
	return math.Round(MBps * 8)
}

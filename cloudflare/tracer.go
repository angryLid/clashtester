package cloudflare

import (
	"fmt"
	"math"
	"net/http/httptrace"
	"time"
)

type Latency time.Duration

func (l Latency) String() string {
	return fmt.Sprintf("%dms", l.Cell())
}

func (l Latency) Cell() int64 {
	r := l / Latency(time.Millisecond)
	return int64(r)
}

type LatencySlice []Latency

func (list LatencySlice) Avg() Latency {
	length := len(list)
	sum := Latency(0)

	if length < 1 {
		return sum
	}

	for _, val := range list {
		sum += val
	}

	return sum / Latency(length)
}

type Speed struct {
	size     int64
	duration time.Duration
}

func (s Speed) Value() float64 {
	if s.duration == 0 {
		return 0
	}
	MB := float64(s.size) / (1024 * 1024)
	sec := float64(s.duration) / float64(time.Second)
	r := MB / sec
	return math.Round(r*100) / 100
}
func (s Speed) Cell() float64 {
	return s.Value()
}
func (s Speed) String() string {
	return fmt.Sprintf("%.2fMB/s", s.Value())
}

type SpeedSlice []Speed

func (list SpeedSlice) Avg() Speed {
	length := len(list)

	sum := Speed{}

	if length < 1 {
		return sum
	}
	var totalWritten int64
	var totalDuration time.Duration
	for _, val := range list {
		totalWritten += val.size
		totalDuration += val.duration
	}

	return Speed{
		size:     totalWritten,
		duration: totalDuration,
	}
}

// Deprecated TODO remove
func (a Speed) Add(b Speed) Speed {
	return Speed{
		size:     a.size + b.size,
		duration: a.duration + b.duration,
	}
}

// Deprecated TODO remove
func NaS() Speed {
	return Speed{
		size:     0,
		duration: time.Second,
	}
}

type Tracer struct {
	onGetConn              time.Time
	onGotConn              time.Time
	onGotFirstResponseByte time.Time
	onGotResponseBody      time.Time
	written                int64
	DebugReused            bool
}

func (t *Tracer) GetConn(hostPort string) {
	t.onGetConn = time.Now()
}

func (t *Tracer) GotConn(gci httptrace.GotConnInfo) {
	t.onGotConn = time.Now()
	t.DebugReused = gci.Reused
}

func (t *Tracer) GotFirstResponseByte() {
	t.onGotFirstResponseByte = time.Now()
}

func (t *Tracer) GotResponseBody(written int64) {
	t.onGotResponseBody = time.Now()
	t.written = written
}

func (t *Tracer) Latency() Latency {
	return Latency(t.onGotFirstResponseByte.Sub(t.onGotConn))
}

func (t *Tracer) Speed() Speed {
	return Speed{
		t.written,
		t.onGotResponseBody.Sub(t.onGotConn),
	}
}

type TracerList []*Tracer

// TODO: remove
func (r TracerList) AvgLatency() Latency {
	var l Latency
	for _, val := range r {
		l += val.Latency()
	}
	if len(r) < 1 {
		return Latency(0)
	}
	return l / Latency(len(r))
}

// TODO remove
func (r TracerList) AvgSpeed() Speed {
	var totalWritten int64
	var totalDuration time.Duration
	for _, val := range r {
		totalWritten += val.Speed().size
		totalDuration += val.Speed().duration
	}

	return Speed{
		size:     totalWritten / int64(len(r)),
		duration: totalDuration / time.Duration(len(r)),
	}
}

type Record struct {
	NodeName string
	Latency  LatencySlice
	Up100KB  SpeedSlice
	Dn100KB  SpeedSlice
	Dn10MB   SpeedSlice
}

func (r Record) String() string {
	return r.NodeName + " { " +
		"Latency " + r.Latency.Avg().String() +
		", Dn100KB " + r.Dn100KB.Avg().String() +
		", Dn10MB " + r.Dn10MB.Avg().String() +
		", Up100KB " + r.Up100KB.Avg().String() + " }"
}

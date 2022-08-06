package cloudflare

import (
	"fmt"
	"net/http/httptrace"
	"time"
)

type Latency time.Duration

func (l Latency) String() string {
	r := l / Latency(time.Millisecond)
	return fmt.Sprintf("%dms", r)
}

type Speed struct {
	size     int64
	duration time.Duration
}

func (s Speed) Value() float64 {
	MB := float64(s.size) / (1024 * 1024)
	sec := float64(s.duration) / float64(time.Second)
	return MB / sec
}

func (s Speed) String() string {
	return fmt.Sprintf("%.2fMB/s", s.Value())
}

func (a Speed) Add(b Speed) Speed {
	return Speed{
		size:     a.size + b.size,
		duration: a.duration + b.duration,
	}
}

// NaS stand for Not a Speed.
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
	Latency Latency
	Up100KB Speed
	Dn100KB Speed
	Dn25MB  Speed
}

func (r Record) String() string {
	return "Latency " + r.Latency.String() +
		", Dn100KB " + r.Dn100KB.String() +
		", Dn25MB " + r.Dn25MB.String() +
		", Up100KB " + r.Up100KB.String()
}

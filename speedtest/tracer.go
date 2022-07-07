package speedtest

import (
	"net/http/httptrace"
	"time"
)

type Tracer struct {
	start        time.Time
	connected    time.Time
	end          time.Time
	totalLatency time.Duration
	latency      time.Duration
}

func (t *Tracer) GetConn(hostPort string) {
	t.start = time.Now()
}
func (t *Tracer) GotConn(gci httptrace.GotConnInfo) {
	t.connected = time.Now()
}
func (t *Tracer) GotFirstResponseByte() {
	t.end = time.Now()
	t.totalLatency = t.end.Sub(t.start)
	t.latency = t.end.Sub(t.connected)
}

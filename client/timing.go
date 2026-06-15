package client

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"sync"
	"time"
)

// Timing captures how long each phase of an HTTP request took. It is populated
// only when Options.Timing is set.
//
// ContentTransfer and Total are intentionally left out here: the client returns
// as soon as the response headers are available, before the body has been read,
// so those two phases can only be known by the caller once it has finished
// consuming the body. The caller derives them from TimeToFirstByte:
//
//	Total           = total elapsed once the body is fully read
//	ContentTransfer = Total - TimeToFirstByte
type Timing struct {
	DNSLookup        time.Duration // resolving the hostname
	TCPConnection    time.Duration // establishing the TCP connection
	TLSHandshake     time.Duration // TLS negotiation (zero for plain http or a reused connection)
	ServerProcessing time.Duration // request written -> first response byte
	TimeToFirstByte  time.Duration // request start -> first response byte
}

// timingKey is the context key under which a timingCollector is carried, so the
// custom DialContext can record the DNS phase that httptrace cannot observe for
// this client (resolution happens in our own concurrent resolver, not the Go
// transport's default resolver).
type timingKey struct{}

func withTiming(ctx context.Context, tc *timingCollector) context.Context {
	return context.WithValue(ctx, timingKey{}, tc)
}

func timingFrom(ctx context.Context) *timingCollector {
	tc, _ := ctx.Value(timingKey{}).(*timingCollector)
	return tc
}

// timingCollector records timestamps for each request phase. All marks use
// "latest wins" semantics so that, across a redirect chain, the durations
// reflect the connection that actually served each phase (a reused keep-alive
// connection simply never re-fires ConnectStart/TLSHandshakeStart, leaving the
// original establishment timings in place).
type timingCollector struct {
	mu           sync.Mutex
	start        time.Time
	dnsStart     time.Time
	dnsDone      time.Time
	connectStart time.Time
	connectDone  time.Time
	tlsStart     time.Time
	tlsDone      time.Time
	wroteRequest time.Time
	firstByte    time.Time
}

func newTimingCollector() *timingCollector {
	return &timingCollector{start: time.Now()}
}

func (tc *timingCollector) mark(field *time.Time) {
	tc.mu.Lock()
	*field = time.Now()
	tc.mu.Unlock()
}

// clientTrace wires the phases the Go transport can observe directly: the TCP
// connect, the TLS handshake, the moment the request was written, and the first
// response byte. DNS is handled separately in DialContext (see markDNS*).
func (tc *timingCollector) clientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		ConnectStart:         func(_, _ string) { tc.mark(&tc.connectStart) },
		ConnectDone:          func(_, _ string, _ error) { tc.mark(&tc.connectDone) },
		TLSHandshakeStart:    func() { tc.mark(&tc.tlsStart) },
		TLSHandshakeDone:     func(tls.ConnectionState, error) { tc.mark(&tc.tlsDone) },
		WroteRequest:         func(httptrace.WroteRequestInfo) { tc.mark(&tc.wroteRequest) },
		GotFirstResponseByte: func() { tc.mark(&tc.firstByte) },
	}
}

func (tc *timingCollector) markDNSStart() { tc.mark(&tc.dnsStart) }
func (tc *timingCollector) markDNSDone()  { tc.mark(&tc.dnsDone) }

func (tc *timingCollector) result() *Timing {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return &Timing{
		DNSLookup:        phase(tc.dnsStart, tc.dnsDone),
		TCPConnection:    phase(tc.connectStart, tc.connectDone),
		TLSHandshake:     phase(tc.tlsStart, tc.tlsDone),
		ServerProcessing: phase(tc.wroteRequest, tc.firstByte),
		TimeToFirstByte:  phase(tc.start, tc.firstByte),
	}
}

// phase returns b-a, or zero if either endpoint was never recorded or the
// timestamps are out of order (e.g. a reused connection that skipped a phase).
func phase(a, b time.Time) time.Duration {
	if a.IsZero() || b.IsZero() || b.Before(a) {
		return 0
	}
	return b.Sub(a)
}

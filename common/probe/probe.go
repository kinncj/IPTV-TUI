// Package probe checks whether a stream URL is reachable. It is a best-effort
// signal for the UI (✓ / ✗), not a guarantee the stream will play — geo-blocks
// and auth walls can pass a probe yet fail in the player, and vice versa.
package probe

import (
	"context"
	"net/http"
	"time"
)

// Status is the reachability verdict for a stream.
type Status int

const (
	Unknown Status = iota // not probed yet
	Checking
	OK      // 2xx / valid manifest response
	Blocked // reachable host but refused (403 / 401) — often geo-blocked
	Dead    // timeout, DNS failure, 4xx/5xx
)

func (s Status) Icon() string {
	switch s {
	case Checking:
		return "…"
	case OK:
		return "✓"
	case Blocked:
		return "⚠"
	case Dead:
		return "✗"
	default:
		return " "
	}
}

// Prober performs bounded-concurrency reachability checks.
type Prober struct {
	client *http.Client
	sem    chan struct{}
}

// New returns a Prober allowing at most `concurrency` in-flight checks, each
// bounded by `timeout`.
func New(concurrency int, timeout time.Duration) *Prober {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Prober{
		client: &http.Client{
			Timeout: timeout,
			// Don't chase redirects forever; a 3xx still means the host answered.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		sem: make(chan struct{}, concurrency),
	}
}

// Check probes a single URL, blocking to respect the concurrency limit.
func (p *Prober) Check(ctx context.Context, url string) Status {
	p.sem <- struct{}{}
	defer func() { <-p.sem }()

	// A ranged GET fetches only the first bytes of an HLS manifest — cheap and
	// more reliable than HEAD, which many CDNs reject.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Dead
	}
	req.Header.Set("Range", "bytes=0-1")
	req.Header.Set("User-Agent", "VLC/3.0 LibVLC/3.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return Dead
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 400:
		return OK
	case resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized:
		return Blocked
	default:
		return Dead
	}
}

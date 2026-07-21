package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// Event is a single request going through the proxy.
// It is published to a channel and consumed by whichever UI is active.
type Event struct {
	Method   string
	Path     string
	Status   int
	Duration time.Duration
	Time     time.Time
}

// newProxyHandler returns an http.Handler that reverse-proxies to the given target
// and publishes an Event for each request to the events channel.
// events may be nil (drops events on the floor).
func newProxyHandler(targetPort int, events chan<- Event) http.Handler {
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", targetPort))
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Preserve client's Host header — dev servers (Vite HMR, Next) check it.
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		host := r.Host
		originalDirector(r)
		r.Host = host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "upstream unreachable at localhost:%d\n", targetPort)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		proxy.ServeHTTP(sw, r)
		if events != nil {
			// Non-blocking send — never let a slow UI stall the proxy.
			select {
			case events <- Event{
				Method:   r.Method,
				Path:     r.URL.Path,
				Status:   sw.status,
				Duration: time.Since(start),
				Time:     start,
			}:
			default:
			}
		}
	})
}

// statusWriter captures the status code from an http.ResponseWriter.
// Also implements http.Hijacker so WebSocket upgrades work.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := s.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijacking not supported")
}

func (s *statusWriter) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

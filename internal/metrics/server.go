package metrics

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ListenAndServe starts an HTTP server on addr serving /metrics from the given
// registry. It validates the address eagerly: if the listener cannot be
// created, it returns an error immediately (fail-fast). Once listening, the
// server runs in a background goroutine; errors after startup are logged but
// do not terminate the process.
func ListenAndServe(addr string, reg *prometheus.Registry) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("metrics listen %s: %w", addr, err)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	srv := &http.Server{Handler: mux}
	log.Printf("metrics server listening on %s", addr)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("metrics server error: %v", err)
		}
	}()
	return nil
}

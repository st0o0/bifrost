//go:build linux

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/st0o0/bifrost/internal/config"
	"github.com/st0o0/bifrost/internal/metrics"
	"github.com/st0o0/bifrost/internal/probe"
	"github.com/st0o0/bifrost/internal/supervisor"
	"github.com/st0o0/bifrost/internal/wg"
)

func main() {
	log.SetPrefix("bifrost: ")
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	if len(os.Args) > 1 {
		if handled, code := runCmd(os.Args[1], os.Stdin, os.Stdout); handled {
			os.Exit(code)
		}
		switch os.Args[1] {
		case "healthcheck":
			os.Exit(runHealthcheck())
		case "handshake":
			os.Exit(runHandshake())
		}
	}

	s, err := config.LoadSettings(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}

	cfg, fromEnv, err := config.LoadConfig(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	if !fromEnv {
		confPath := "/etc/wireguard/" + s.Interface + ".conf"
		f, err := os.Open(confPath)
		if err != nil {
			log.Fatalf("config not found at %s — mount your WireGuard .conf there or set BIFROST_PRIVATE_KEY", confPath)
		}
		cfg, err = config.ParseConfig(f)
		f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	tun, err := wg.Bring(cfg, s.Interface)
	if err != nil {
		log.Fatalf("bringing up %s: %v", s.Interface, err)
	}
	defer tun.Close()

	var stats *metrics.Stats
	if s.Metrics {
		stats = &metrics.Stats{}
		tun.OnEndpointChange = stats.IncEndpointChanges

		collector := metrics.NewCollector(metrics.CollectorOpts{
			Stats:      stats,
			Device:     tun.Device,
			StaleAfter: s.StaleAfter,
			ProbeOn:    s.Probe,
		})
		reg := prometheus.NewRegistry()
		reg.MustRegister(collector)
		if err := metrics.ListenAndServe(s.MetricsAddr, reg); err != nil {
			log.Fatal(err)
		}
	}

	if !s.Resolve && !s.Reconnect {
		log.Print("recovery disabled (resolve and reconnect both off)")
		<-ctx.Done()
		return
	}
	supervisor.Run(ctx, supervisor.Deps{Ctrl: tun, Pinger: probe.ICMPPinger{}, Stats: stats}, s, probe.Targets(cfg, s.ProbeHost))
}

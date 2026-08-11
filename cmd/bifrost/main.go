//go:build linux

package main

import (
	"context"
	"log/slog"
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
		slog.Error("failed to load settings", "error", err)
		os.Exit(1)
	}
	slog.SetDefault(newLogger(s))

	cfg, fromEnv, err := config.LoadConfig(os.Getenv)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if !fromEnv {
		confPath := "/etc/wireguard/" + s.Interface + ".conf"
		f, err := os.Open(confPath)
		if err != nil {
			slog.Error("config not found", "path", confPath)
			os.Exit(1)
		}
		cfg, err = config.ParseConfig(f)
		f.Close()
		if err != nil {
			slog.Error("failed to parse config", "error", err)
			os.Exit(1)
		}
	}
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	tun, err := wg.Bring(cfg, s.Interface)
	if err != nil {
		slog.Error("failed to bring up interface", "interface", s.Interface, "error", err)
		os.Exit(1)
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
			slog.Error("failed to start metrics server", "error", err)
			os.Exit(1)
		}
	}

	if !s.Resolve && !s.Reconnect {
		slog.Info("recovery disabled")
		<-ctx.Done()
		return
	}
	supervisor.Run(ctx, supervisor.Deps{Ctrl: tun, Pinger: probe.ICMPPinger{}, Stats: stats}, s, probe.Targets(cfg, s.ProbeHost))
}

func newLogger(s *config.Settings) *slog.Logger {
	opts := &slog.HandlerOptions{Level: s.LogLevel}
	if s.LogFormat == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

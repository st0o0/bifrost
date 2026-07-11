//go:build linux

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/st0o0/bifrost/internal/config"
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
	confPath := "/etc/wireguard/" + s.Interface + ".conf"
	f, err := os.Open(confPath)
	if err != nil {
		log.Fatalf("config not found at %s — mount your WireGuard .conf there", confPath)
	}
	cfg, err := config.ParseConfig(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
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

	if !s.Resolve && !s.Reconnect {
		log.Print("recovery disabled (resolve and reconnect both off)")
		<-ctx.Done()
		return
	}
	supervisor.Run(ctx, supervisor.Deps{Ctrl: tun, Pinger: probe.ICMPPinger{}}, s, probe.Targets(cfg, s.ProbeHost))
}

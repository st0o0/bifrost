//go:build linux

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/st0o0/bifrost/internal/config"
	"github.com/st0o0/bifrost/internal/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func runHealthcheck() int {
	s, err := config.LoadSettings(os.Getenv)
	if err != nil {
		return 1
	}
	if !s.Healthcheck {
		return 0 // disabled -> always healthy
	}
	c, err := wgctrl.New()
	if err != nil {
		return 1
	}
	defer c.Close()
	newest := wg.NewestHandshakeVia(c, s.Interface)
	if healthy(newest, time.Now(), s.HealthStaleAfter) {
		return 0
	}
	fmt.Fprintln(os.Stderr, "bifrost: tunnel unhealthy")
	return 1
}

//go:build linux

package main

import (
	"fmt"
	"os"

	"github.com/st0o0/bifrost/internal/config"
	"github.com/st0o0/bifrost/internal/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
)

// runHandshake prints the newest peer handshake as a unix epoch (0 if none)
// for the configured interface, so a scratch-image e2e test can query the
// tunnel's handshake state without a shell.
func runHandshake() int {
	s, err := config.LoadSettings(os.Getenv)
	if err != nil {
		fmt.Println(0)
		return 0
	}
	c, err := wgctrl.New()
	if err != nil {
		fmt.Println(0)
		return 0
	}
	defer c.Close()
	newest := wg.NewestHandshakeVia(c, s.Interface)
	if newest.IsZero() {
		fmt.Println(0)
		return 0
	}
	fmt.Println(newest.Unix())
	return 0
}

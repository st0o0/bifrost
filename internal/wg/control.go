//go:build linux

package wg

import (
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/st0o0/bifrost/internal/recovery"
)

// var _ ensures Tunnel keeps satisfying recovery.Controller.
var _ recovery.Controller = (*Tunnel)(nil)

// NewestHandshake returns the newest LastHandshakeTime across peers (zero if
// none yet).
func (t *Tunnel) NewestHandshake() time.Time {
	return NewestHandshakeVia(t.ctrl, t.iface)
}

// NewestHandshakeVia queries an already-open wgctrl client for the newest
// handshake on iface. It's a standalone helper (rather than a Tunnel method)
// so a short-lived process — the `bifrost healthcheck` subcommand — can query
// an already-configured interface without going through Bring() (which would
// try to create the link and reconfigure it).
func NewestHandshakeVia(client *wgctrl.Client, iface string) time.Time {
	dev, err := client.Device(iface)
	if err != nil {
		return time.Time{}
	}
	var newest time.Time
	for _, p := range dev.Peers {
		if p.LastHandshakeTime.After(newest) {
			newest = p.LastHandshakeTime
		}
	}
	return newest
}

// Resolve re-resolves each peer's hostname endpoint and updates it in place
// (UpdateOnly: true — never creates a new peer). Peers with no hostname
// endpoint, or whose hostname doesn't currently resolve, are left untouched.
func (t *Tunnel) Resolve() error {
	var peers []wgtypes.PeerConfig
	for _, p := range t.cfg.Peers {
		if p.EndpointHost == "" {
			continue
		}
		ep, err := resolveUDP(p.EndpointHost, p.EndpointPort)
		if err != nil {
			continue
		}
		key, err := wgtypes.ParseKey(p.PublicKey)
		if err != nil {
			continue
		}
		peers = append(peers, wgtypes.PeerConfig{PublicKey: key, Endpoint: ep, UpdateOnly: true})
	}
	if len(peers) == 0 {
		return nil
	}
	return t.ctrl.ConfigureDevice(t.iface, wgtypes.Config{Peers: peers})
}

// Reconnect tears the interface down and brings it back up from scratch
// (re-resolving endpoints and reapplying config in the process).
func (t *Tunnel) Reconnect() error {
	if err := t.Close(); err != nil {
		// Best-effort teardown; keep going and try to recreate anyway.
		_ = err
	}
	nt, err := Bring(t.cfg, t.iface)
	if err != nil {
		return err
	}
	*t = *nt
	return nil
}

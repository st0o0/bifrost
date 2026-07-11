//go:build linux

// Package wg brings up a real WireGuard interface: kernel WireGuard first
// (via netlink), falling back to the embedded userspace implementation
// (wireguard-go) when the kernel module isn't available. Configuration
// (keys/peers) is applied via wgctrl; addresses and split-tunnel routes are
// applied via netlink.
package wg

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/st0o0/bifrost/internal/config"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Tunnel is a live WireGuard interface plus the info needed to re-resolve and
// rebuild it.
type Tunnel struct {
	iface     string
	cfg       *config.Config
	userspace bool
	dev       *device.Device // non-nil in userspace mode
	uapi      net.Listener   // non-nil in userspace mode
	ctrl      *wgctrl.Client
}

// Bring creates and configures the interface: kernel WireGuard first, falling
// back to the embedded userspace implementation; then applies keys/peers via
// wgctrl and sets up addresses/split-tunnel routes, and brings the link up.
func Bring(cfg *config.Config, iface string) (t *Tunnel, retErr error) {
	t = &Tunnel{iface: iface, cfg: cfg}
	if err := t.createLink(); err != nil {
		return nil, fmt.Errorf("bifrost/wg: create link %s: %w", iface, err)
	}
	defer func() {
		if retErr != nil {
			_ = t.Close()
		}
	}()
	ctrl, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("bifrost/wg: open wgctrl: %w", err)
	}
	t.ctrl = ctrl
	if err := t.configure(); err != nil {
		return nil, fmt.Errorf("bifrost/wg: configure %s: %w", iface, err)
	}
	if err := t.setupNetwork(); err != nil {
		return nil, fmt.Errorf("bifrost/wg: setup network %s: %w", iface, err)
	}
	return t, nil
}

// createLink brings the network interface into existence: kernel WireGuard
// (via netlink) first, falling back to the embedded userspace implementation
// (TUN device + UAPI control socket) if the kernel module isn't available.
func (t *Tunnel) createLink() error {
	la := netlink.NewLinkAttrs()
	la.Name = t.iface
	err := netlink.LinkAdd(&netlink.Wireguard{LinkAttrs: la})
	if err == nil {
		log.Printf("bifrost/wg: %s: using kernel WireGuard", t.iface)
		return nil
	}
	if isExist(err) {
		// Leftover link from a crashed prior run: delete it and retry once,
		// rather than falling back to userspace.
		log.Printf("bifrost/wg: %s: link already exists, deleting stale link and retrying", t.iface)
		if stale, lerr := netlink.LinkByName(t.iface); lerr == nil {
			if derr := netlink.LinkDel(stale); derr != nil {
				return fmt.Errorf("delete stale link %s: %w", t.iface, derr)
			}
		}
		if err = netlink.LinkAdd(&netlink.Wireguard{LinkAttrs: la}); err == nil {
			log.Printf("bifrost/wg: %s: using kernel WireGuard", t.iface)
			return nil
		}
	}
	log.Printf("bifrost/wg: %s: kernel WireGuard unavailable (%v), falling back to userspace", t.iface, err)
	// Kernel WireGuard unavailable; fall back to the userspace implementation.

	tdev, err := tun.CreateTUN(t.iface, mtuOr(t.cfg, 1420))
	if err != nil {
		return fmt.Errorf("create TUN device: %w", err)
	}
	dev := device.NewDevice(tdev, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "bifrost/wg: "))
	t.dev = dev

	f, err := ipc.UAPIOpen(t.iface)
	if err != nil {
		dev.Close()
		return fmt.Errorf("open UAPI socket: %w", err)
	}
	uapi, err := ipc.UAPIListen(t.iface, f)
	if err != nil {
		dev.Close()
		return fmt.Errorf("listen on UAPI socket: %w", err)
	}
	t.uapi = uapi

	go func() {
		for {
			c, err := uapi.Accept()
			if err != nil {
				// Listener closed (Close()); stop accepting.
				return
			}
			go dev.IpcHandle(c)
		}
	}()

	if err := dev.Up(); err != nil {
		_ = uapi.Close()
		dev.Close()
		return fmt.Errorf("bring up userspace device: %w", err)
	}

	t.userspace = true
	log.Printf("bifrost/wg: %s: using userspace WireGuard (wireguard-go)", t.iface)
	return nil
}

// configure applies the private key, listen port, and peer list via wgctrl.
// Peer endpoints that don't resolve yet are left unset; Resolve() fills them
// in once DNS answers.
func (t *Tunnel) configure() error {
	priv, err := wgtypes.ParseKey(t.cfg.PrivateKey)
	if err != nil {
		return fmt.Errorf("parse private key: %w", err)
	}

	cfg := wgtypes.Config{
		PrivateKey:   &priv,
		ReplacePeers: true,
	}
	if t.cfg.ListenPort != 0 {
		port := t.cfg.ListenPort
		cfg.ListenPort = &port
	}

	for _, p := range t.cfg.Peers {
		pub, err := wgtypes.ParseKey(p.PublicKey)
		if err != nil {
			return fmt.Errorf("parse peer public key %q: %w", p.PublicKey, err)
		}

		pc := wgtypes.PeerConfig{
			PublicKey:         pub,
			ReplaceAllowedIPs: true,
			AllowedIPs:        toIPNets(p.AllowedIPs),
		}

		if p.PresharedKey != "" {
			psk, err := wgtypes.ParseKey(p.PresharedKey)
			if err != nil {
				return fmt.Errorf("parse peer preshared key: %w", err)
			}
			pc.PresharedKey = &psk
		}

		if p.PersistentKeepalive != 0 {
			d := time.Duration(p.PersistentKeepalive) * time.Second
			pc.PersistentKeepaliveInterval = &d
		}

		if p.EndpointHost != "" {
			ep, err := resolveUDP(p.EndpointHost, p.EndpointPort)
			if err != nil {
				log.Printf("bifrost/wg: %s: peer %s endpoint %s:%d does not resolve yet: %v", t.iface, pub, p.EndpointHost, p.EndpointPort, err)
			} else {
				pc.Endpoint = ep
			}
		}

		cfg.Peers = append(cfg.Peers, pc)
	}

	return t.ctrl.ConfigureDevice(t.iface, cfg)
}

// setupNetwork assigns the interface's addresses, MTU, and split-tunnel
// routes, then brings the link up. Default routes (0.0.0.0/0, ::/0) are
// skipped with a warning — full-tunnel routing is deferred. DNS= is ignored
// entirely (warned by the caller during config parsing).
func (t *Tunnel) setupNetwork() error {
	link, err := netlink.LinkByName(t.iface)
	if err != nil {
		return fmt.Errorf("lookup link %s: %w", t.iface, err)
	}

	for _, a := range t.cfg.Addresses {
		addr, err := netlink.ParseAddr(a.String())
		if err != nil {
			return fmt.Errorf("parse address %s: %w", a, err)
		}
		if err := netlink.AddrAdd(link, addr); err != nil {
			if isExist(err) {
				log.Printf("bifrost/wg: %s: address %s already exists, ignoring", t.iface, a)
				continue
			}
			return fmt.Errorf("add address %s: %w", a, err)
		}
	}

	if t.cfg.MTU != 0 {
		if err := netlink.LinkSetMTU(link, t.cfg.MTU); err != nil {
			return fmt.Errorf("set MTU %d: %w", t.cfg.MTU, err)
		}
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("set link up: %w", err)
	}

	for _, p := range t.cfg.Peers {
		for _, prefix := range p.AllowedIPs {
			if isDefaultRoute(prefix) {
				log.Printf("bifrost/wg: %s: skipping full-tunnel route %s (full-tunnel deferred)", t.iface, prefix)
				continue
			}
			route := &netlink.Route{
				LinkIndex: link.Attrs().Index,
				Dst:       toIPNet(prefix),
				Scope:     netlink.SCOPE_LINK,
			}
			if err := netlink.RouteAdd(route); err != nil {
				if isExist(err) {
					log.Printf("bifrost/wg: %s: route %s already exists, ignoring", t.iface, prefix)
					continue
				}
				return fmt.Errorf("add route %s: %w", prefix, err)
			}
		}
	}

	return nil
}

// Close tears the interface down: deletes the netlink device (which also
// works for kernel-created wireguard links) and releases the userspace
// device/UAPI socket and wgctrl client, if any.
func (t *Tunnel) Close() error {
	var errs []error

	if link, err := netlink.LinkByName(t.iface); err == nil {
		if err := netlink.LinkDel(link); err != nil {
			errs = append(errs, fmt.Errorf("delete link %s: %w", t.iface, err))
		}
	} else if !errors.As(err, new(netlink.LinkNotFoundError)) {
		errs = append(errs, fmt.Errorf("lookup link %s: %w", t.iface, err))
	}

	if t.userspace {
		if t.uapi != nil {
			if err := t.uapi.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close UAPI listener: %w", err))
			}
		}
		if t.dev != nil {
			t.dev.Close()
		}
	}

	if t.ctrl != nil {
		if err := t.ctrl.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close wgctrl client: %w", err))
		}
	}

	return errors.Join(errs...)
}

// isExist reports whether err indicates the target of a netlink operation
// (link, address, or route) already exists — i.e. it's safe to treat as
// non-fatal rather than aborting bring-up.
func isExist(err error) bool {
	return errors.Is(err, os.ErrExist) || errors.Is(err, syscall.EEXIST)
}

// isDefaultRoute reports whether prefix is a default route (0.0.0.0/0 or
// ::/0) — full-tunnel routing, deferred to a later phase.
func isDefaultRoute(p netip.Prefix) bool {
	return p.Bits() == 0
}

// toIPNets converts a list of netip.Prefix to net.IPNet, as wgtypes.PeerConfig
// expects.
func toIPNets(ps []netip.Prefix) []net.IPNet {
	out := make([]net.IPNet, 0, len(ps))
	for _, p := range ps {
		out = append(out, *toIPNet(p))
	}
	return out
}

// toIPNet converts a single netip.Prefix to *net.IPNet.
func toIPNet(p netip.Prefix) *net.IPNet {
	addr := p.Addr()
	return &net.IPNet{
		IP:   addr.AsSlice(),
		Mask: net.CIDRMask(p.Bits(), addr.BitLen()),
	}
}

// resolveUDP resolves host:port to a *net.UDPAddr, re-resolving hostnames
// (including DDNS ones) on every call.
func resolveUDP(host string, port int) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.Itoa(port)))
}

// mtuOr returns cfg.MTU if set, else def.
func mtuOr(cfg *config.Config, def int) int {
	if cfg != nil && cfg.MTU != 0 {
		return cfg.MTU
	}
	return def
}

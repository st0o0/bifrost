package main

import (
	"fmt"
	"io"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

func pubKey(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	k, err := wgtypes.ParseKey(strings.TrimSpace(string(b)))
	if err != nil {
		return "", err
	}
	return k.PublicKey().String(), nil
}

// runCmd handles the platform-independent subcommands (genkey, pubkey,
// version); returns (handled, exitCode).
func runCmd(cmd string, stdin io.Reader, stdout io.Writer) (bool, int) {
	switch cmd {
	case "genkey":
		k, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return true, 1
		}
		fmt.Fprintln(stdout, k.String())
		return true, 0
	case "pubkey":
		p, err := pubKey(stdin)
		if err != nil {
			return true, 1
		}
		fmt.Fprintln(stdout, p)
		return true, 0
	case "version":
		fmt.Fprintln(stdout, version)
		return true, 0
	}
	return false, 0
}

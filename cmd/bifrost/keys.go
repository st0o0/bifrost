package main

import (
	"fmt"
	"io"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func genKey() string {
	k, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return ""
	}
	return k.String()
}

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

// runKeyCmd handles the genkey/pubkey subcommands; returns (handled, exitCode).
func runKeyCmd(cmd string, stdin io.Reader, stdout io.Writer) (bool, int) {
	switch cmd {
	case "genkey":
		fmt.Fprintln(stdout, genKey())
		return true, 0
	case "pubkey":
		p, err := pubKey(stdin)
		if err != nil {
			return true, 1
		}
		fmt.Fprintln(stdout, p)
		return true, 0
	}
	return false, 0
}

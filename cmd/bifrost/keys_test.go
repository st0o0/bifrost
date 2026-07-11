package main

import (
	"strings"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestGenAndPubKey(t *testing.T) {
	var priv strings.Builder
	if handled, code := runCmd("genkey", nil, &priv); !handled || code != 0 {
		t.Fatalf("genkey: handled=%v code=%d", handled, code)
	}
	if _, err := wgtypes.ParseKey(strings.TrimSpace(priv.String())); err != nil {
		t.Fatalf("genkey not a valid key: %v", err)
	}

	var pub strings.Builder
	if handled, code := runCmd("pubkey", strings.NewReader(priv.String()), &pub); !handled || code != 0 {
		t.Fatalf("pubkey: handled=%v code=%d", handled, code)
	}
	if _, err := wgtypes.ParseKey(strings.TrimSpace(pub.String())); err != nil {
		t.Fatalf("pubkey not a valid key: %v", err)
	}
	if strings.TrimSpace(pub.String()) == strings.TrimSpace(priv.String()) {
		t.Error("public key equals private key")
	}
}

func TestPubKeyRejectsGarbage(t *testing.T) {
	if _, code := runCmd("pubkey", strings.NewReader("not a key"), &strings.Builder{}); code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
}

func TestUnknownCmdNotHandled(t *testing.T) {
	if handled, _ := runCmd("frobnicate", nil, &strings.Builder{}); handled {
		t.Error("unknown command reported as handled")
	}
}

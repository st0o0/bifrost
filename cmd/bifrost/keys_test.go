package main

import (
	"strings"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestGenAndPubKey(t *testing.T) {
	priv := genKey()
	if _, err := wgtypes.ParseKey(strings.TrimSpace(priv)); err != nil {
		t.Fatalf("genKey not a valid key: %v", err)
	}
	pub, err := pubKey(strings.NewReader(priv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wgtypes.ParseKey(strings.TrimSpace(pub)); err != nil {
		t.Fatalf("pubKey not a valid key: %v", err)
	}
	if strings.TrimSpace(pub) == strings.TrimSpace(priv) {
		t.Error("public key equals private key")
	}
}

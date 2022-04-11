package main

import (
	"testing"

	"github.com/quay/claircore"
)

func TestDigests(t *testing.T) {
	_, err := claircore.ParseDigest("sha256:ce419f530850a57ffb757cd9dfc06b6645124d63f1c15459ec799d0555815f77 ")
	if err != nil {
		t.Fatalf("could not create digest: %v", err)
	}
}

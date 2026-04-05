package security

import (
	"strings"
	"testing"
)

func TestFingerprintFormatAndStability(t *testing.T) {
	pub := make([]byte, 32)
	for i := range pub {
		pub[i] = byte(i)
	}
	fp := FingerprintOf(pub)
	if !strings.HasPrefix(fp, "SHA256:") {
		t.Fatalf("fingerprint missing SHA256 prefix: %q", fp)
	}
	if fp != FingerprintOf(pub) {
		t.Fatal("fingerprint is not deterministic")
	}

	other := make([]byte, 32)
	other[0] = 0xff
	if fp == FingerprintOf(other) {
		t.Fatal("different keys produced the same fingerprint")
	}
}

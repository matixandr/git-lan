package security

import (
	"errors"
	"testing"
)

func TestTrustVerifyHost(t *testing.T) {
	tr := &TrustRing{Entries: map[string]TrustEntry{}}
	tr.Add("maciek-laptop", "SHA256:abc123")

	// Known + matching fingerprint → trusted.
	trusted, err := tr.VerifyHost("maciek-laptop", "SHA256:abc123")
	if err != nil || !trusted {
		t.Fatalf("expected trusted, got trusted=%v err=%v", trusted, err)
	}

	// Known + mismatching fingerprint → MITM error.
	_, err = tr.VerifyHost("maciek-laptop", "SHA256:evil999")
	if !errors.Is(err, ErrFingerprintMismatch) {
		t.Fatalf("expected ErrFingerprintMismatch, got %v", err)
	}

	// Unknown host → not trusted, no error (TOFU territory).
	trusted, err = tr.VerifyHost("stranger", "SHA256:whatever")
	if err != nil || trusted {
		t.Fatalf("expected untrusted+no error, got trusted=%v err=%v", trusted, err)
	}
}

func TestTrustAddRemove(t *testing.T) {
	tr := &TrustRing{Entries: map[string]TrustEntry{}}
	tr.Add("host", "SHA256:fp")
	if _, ok := tr.Get("host"); !ok {
		t.Fatal("entry not added")
	}
	if !tr.Remove("host") {
		t.Fatal("remove returned false for existing entry")
	}
	if tr.Remove("host") {
		t.Fatal("remove returned true for missing entry")
	}
}

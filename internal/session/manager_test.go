package session

import (
	"testing"
	"time"
)

func TestSessionPasswordPolicy(t *testing.T) {
	open, err := New("jam", "", "/repo", true)
	if err != nil {
		t.Fatal(err)
	}
	if open.HasPassword() {
		t.Error("empty password should leave session open")
	}
	if !open.CheckPassword("anything") {
		t.Error("open session should accept any password")
	}

	locked, _ := New("secret-jam", "hunter2", "/repo", false)
	if !locked.HasPassword() {
		t.Error("expected password-protected session")
	}
	if locked.CheckPassword("wrong") {
		t.Error("wrong password accepted")
	}
	if !locked.CheckPassword("hunter2") {
		t.Error("correct password rejected")
	}
}

func TestInviteBurnIsOneTime(t *testing.T) {
	s, _ := New("s", "", "/repo", false)
	_, id, _ := GenerateInvite(s.Secret, time.Hour)

	if s.IsBurned(id) {
		t.Fatal("fresh invite reported burned")
	}
	if !s.Burn(id) {
		t.Fatal("first burn should succeed")
	}
	if s.Burn(id) {
		t.Fatal("second burn should fail (one-time)")
	}
	if !s.IsBurned(id) {
		t.Fatal("invite should be burned after use")
	}
}

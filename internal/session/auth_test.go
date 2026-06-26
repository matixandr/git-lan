package session

import "testing"

func TestPasswordHashVerify(t *testing.T) {
	enc, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("correct horse battery staple", enc) {
		t.Error("correct password did not verify")
	}
	if VerifyPassword("wrong", enc) {
		t.Error("wrong password verified")
	}
}

func TestHashesAreSalted(t *testing.T) {
	a, _ := HashPassword("same")
	b, _ := HashPassword("same")
	if a == b {
		t.Error("identical passwords produced identical hashes (missing salt)")
	}
}

func TestVerifyRejectsGarbage(t *testing.T) {
	if VerifyPassword("x", "not-a-phc-string") {
		t.Error("garbage encoding verified")
	}
	if VerifyPassword("x", "$argon2id$v=19$m=1$bad$bad") {
		t.Error("malformed PHC verified")
	}
}

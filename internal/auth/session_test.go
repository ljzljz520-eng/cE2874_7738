package auth

import (
	"testing"
	"time"
)

func TestDirectoryLoginLifecycle(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	d := NewDirectory(func() time.Time { return now }, "test")
	if _, err := d.AddTherapist("t1", "Lin", "Rehab", "secret"); err != nil {
		t.Fatal(err)
	}
	session, err := d.Login("t1", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Authenticate(session.Token); err != nil {
		t.Fatal(err)
	}
	if !d.Logout(session.Token) {
		t.Fatal("expected logout")
	}
	if _, err := d.Authenticate(session.Token); err == nil {
		t.Fatal("expected invalid session")
	}
}

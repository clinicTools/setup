package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionRoundTrip(t *testing.T) {
	mgr := NewSessionManager([]byte("test-secret-0123456789abcdef"), time.Hour, false)
	user := &User{Username: "maria", UID: 1000, Admin: true, Groups: []string{"sudo"}}

	// Cookie ausstellen.
	rec := httptest.NewRecorder()
	if err := mgr.Issue(rec, user); err != nil {
		t.Fatalf("Issue: %v", err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("kein Cookie gesetzt")
	}

	// Cookie verifizieren.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	got, err := mgr.Verify(req)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got.Username != "maria" || !got.Admin {
		t.Errorf("Verify lieferte %+v", got)
	}
}

func TestVerifyRejectsTamperedToken(t *testing.T) {
	mgr := NewSessionManager([]byte("secret-a"), time.Hour, false)
	other := NewSessionManager([]byte("secret-b"), time.Hour, false)

	rec := httptest.NewRecorder()
	_ = mgr.Issue(rec, &User{Username: "x"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}

	// Mit fremdem Secret signiertes Token muss abgelehnt werden.
	if _, err := other.Verify(req); err == nil {
		t.Error("Verify akzeptierte ein mit fremdem Secret signiertes Token")
	}
}

func TestVerifyNoCookie(t *testing.T) {
	mgr := NewSessionManager([]byte("secret"), time.Hour, false)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := mgr.Verify(req); err == nil {
		t.Error("Verify ohne Cookie sollte fehlschlagen")
	}
}

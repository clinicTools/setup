package system

import "testing"

func TestValidUsername(t *testing.T) {
	valid := []string{"root", "maria", "user_1", "a-b.c", "svc-app"}
	for _, name := range valid {
		if !validUsername(name) {
			t.Errorf("validUsername(%q) = false, erwartet true", name)
		}
	}

	invalid := []string{"", "1abc", "-root", "a b", "rm;reboot", "user$", string(make([]byte, 40))}
	for _, name := range invalid {
		if validUsername(name) {
			t.Errorf("validUsername(%q) = true, erwartet false", name)
		}
	}
}

func TestValidUnit(t *testing.T) {
	if err := validUnit("ssh.service"); err != nil {
		t.Errorf("validUnit(ssh.service) unerwarteter Fehler: %v", err)
	}
	for _, bad := range []string{"", "-x", "a;b", "a b", "a|b", "$(x)"} {
		if err := validUnit(bad); err == nil {
			t.Errorf("validUnit(%q) = nil, erwartet Fehler", bad)
		}
	}
}

func TestValidTimezone(t *testing.T) {
	for _, tz := range []string{"Europe/Berlin", "UTC", "America/New_York"} {
		if !validTimezone(tz) {
			t.Errorf("validTimezone(%q) = false, erwartet true", tz)
		}
	}
	for _, tz := range []string{"", "Berlin", "../etc/passwd", "-flag", "a;b"} {
		if validTimezone(tz) {
			t.Errorf("validTimezone(%q) = true, erwartet false", tz)
		}
	}
}

func TestIsSafeToken(t *testing.T) {
	if !isSafeToken("err") || !isSafeToken("warning4") {
		t.Error("erwartete sichere Tokens wurden abgelehnt")
	}
	for _, bad := range []string{"", "a b", "a;b", "../x"} {
		if isSafeToken(bad) {
			t.Errorf("isSafeToken(%q) = true, erwartet false", bad)
		}
	}
}

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

func TestValidPackageName(t *testing.T) {
	for _, p := range []string{"htop", "lib-foo", "g++", "python3.11", "name:amd64", "ca-certificates"} {
		if !ValidPackageName(p) {
			t.Errorf("ValidPackageName(%q) = false, erwartet true", p)
		}
	}
	for _, p := range []string{"", "-rf", "a;b", "a b", "a/b", "$(x)", "a&b"} {
		if ValidPackageName(p) {
			t.Errorf("ValidPackageName(%q) = true, erwartet false", p)
		}
	}
}

func TestValidPort(t *testing.T) {
	for _, p := range []string{"22", "443", "65535", "1"} {
		if !validPort(p) {
			t.Errorf("validPort(%q) = false, erwartet true", p)
		}
	}
	for _, p := range []string{"", "0", "65536", "-1", "abc", "22/tcp"} {
		if validPort(p) {
			t.Errorf("validPort(%q) = true, erwartet false", p)
		}
	}
}

func TestValidCronSchedule(t *testing.T) {
	for _, s := range []string{"0 * * * *", "*/5 0 1,15 * 1-5", "0 0 * * 0"} {
		if !validCronSchedule(s) {
			t.Errorf("validCronSchedule(%q) = false, erwartet true", s)
		}
	}
	for _, s := range []string{"", "0 * * *", "0 * * * * *", "a * * * *", "0;0 * * * *"} {
		if validCronSchedule(s) {
			t.Errorf("validCronSchedule(%q) = true, erwartet false", s)
		}
	}
}

func TestValidHostname(t *testing.T) {
	for _, h := range []string{"server01", "kube-node.local", "a"} {
		if !validHostname(h) {
			t.Errorf("validHostname(%q) = false, erwartet true", h)
		}
	}
	for _, h := range []string{"", "-bad", "a b", "host;rm", "a/b"} {
		if validHostname(h) {
			t.Errorf("validHostname(%q) = true, erwartet false", h)
		}
	}
}

func TestValidNoControl(t *testing.T) {
	for _, s := range []string{"Sicher!2024", "mit spaces", "üml@ut:ok", ""} {
		if !validNoControl(s) {
			t.Errorf("validNoControl(%q) = false, erwartet true", s)
		}
	}
	// Zeilenumbruch/Steuerzeichen müssen abgelehnt werden (chpasswd-Injection).
	for _, s := range []string{"pw\nroot:pwned", "a\rb", "tab\tval", "null\x00"} {
		if validNoControl(s) {
			t.Errorf("validNoControl(%q) = true, erwartet false", s)
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

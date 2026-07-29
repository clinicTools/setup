package system

import "testing"

func TestValidStackName(t *testing.T) {
	for _, n := range []string{"web", "my-stack", "app_1", "nextcloud", "a"} {
		if !validStackName(n) {
			t.Errorf("validStackName(%q) = false, erwartet true", n)
		}
	}
	// Grossbuchstaben sind fuer Compose-Projektnamen unzulaessig, ebenso
	// Sonderzeichen und fuehrende Trennzeichen (Schutz vor Pfad-/Flag-Injection).
	for _, n := range []string{"", "-web", "_x", "My-Stack", "a b", "a/b", "../etc", "a;b", "a$b"} {
		if validStackName(n) {
			t.Errorf("validStackName(%q) = true, erwartet false", n)
		}
	}
}

func TestValidContainerRef(t *testing.T) {
	for _, ref := range []string{"a1b2c3d4e5f6", "web-1", "my_container", "app.svc"} {
		if !validContainerRef(ref) {
			t.Errorf("validContainerRef(%q) = false, erwartet true", ref)
		}
	}
	for _, ref := range []string{"", "-f", "a b", "a;rm -rf /", "a$(x)", "a|b"} {
		if validContainerRef(ref) {
			t.Errorf("validContainerRef(%q) = true, erwartet false", ref)
		}
	}
}

func TestStackPathIsConfined(t *testing.T) {
	// Der Pfad muss immer unterhalb des verwalteten Verzeichnisses liegen.
	got := StackPath("web")
	want := stacksDir + "/web/" + composeFileName
	if got != want {
		t.Errorf("StackPath = %q, erwartet %q", got, want)
	}
}

func TestStackComposeArgsRejectsBadInput(t *testing.T) {
	if _, _, err := StackComposeArgs("../evil", "up"); err == nil {
		t.Error("StackComposeArgs akzeptierte ungueltigen Stack-Namen")
	}
	if _, _, err := StackComposeArgs("web", "rm -rf"); err == nil {
		t.Error("StackComposeArgs akzeptierte ungueltige Aktion")
	}
}

func TestComposeProjectLabels(t *testing.T) {
	if got := composeProject(map[string]string{"com.docker.compose.project": "web"}); got != "web" {
		t.Errorf("composeProject(docker) = %q, erwartet web", got)
	}
	if got := composeProject(map[string]string{"io.podman.compose.project": "api"}); got != "api" {
		t.Errorf("composeProject(podman) = %q, erwartet api", got)
	}
	if got := composeProject(map[string]string{}); got != "" {
		t.Errorf("composeProject(leer) = %q, erwartet \"\"", got)
	}
}

func TestShortID(t *testing.T) {
	if got := shortID("sha256:abcdef0123456789"); got != "abcdef012345" {
		t.Errorf("shortID = %q", got)
	}
	if got := shortID("abc"); got != "abc" {
		t.Errorf("shortID(kurz) = %q", got)
	}
}

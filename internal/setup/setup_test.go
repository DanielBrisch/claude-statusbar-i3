package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func settingsPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "settings.json")
}

func read(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("settings is not valid JSON: %v\n%s", err, b)
	}
	return m
}

func TestInstallCreatesSettingsWhenThereIsNone(t *testing.T) {
	p := settingsPath(t)

	if err := InstallStatusLine(p, "claude-statusbar collect", false); err != nil {
		t.Fatalf("InstallStatusLine: %v", err)
	}

	sl, ok := read(t, p)["statusLine"].(map[string]any)
	if !ok {
		t.Fatal("statusLine missing from settings")
	}
	if got, want := sl["type"], "command"; got != want {
		t.Errorf("type = %v, want %v", got, want)
	}
	if got, want := sl["command"], "claude-statusbar collect"; got != want {
		t.Errorf("command = %v, want %v", got, want)
	}
}

func TestInstallKeepsEverythingElseInSettings(t *testing.T) {
	p := settingsPath(t)
	if err := os.WriteFile(p, []byte(`{"model":"opus","env":{"FOO":"bar"}}`), 0o600); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	if err := InstallStatusLine(p, "claude-statusbar collect", false); err != nil {
		t.Fatalf("InstallStatusLine: %v", err)
	}

	m := read(t, p)
	if got, want := m["model"], "opus"; got != want {
		t.Errorf("model = %v, want %v — unrelated settings must survive", got, want)
	}
	env, ok := m["env"].(map[string]any)
	if !ok || env["FOO"] != "bar" {
		t.Errorf("env = %v, want it preserved", m["env"])
	}
}

func TestInstallBacksUpBeforeTouchingAnExistingFile(t *testing.T) {
	p := settingsPath(t)
	original := `{"model":"opus"}`
	if err := os.WriteFile(p, []byte(original), 0o600); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	if err := InstallStatusLine(p, "claude-statusbar collect", false); err != nil {
		t.Fatalf("InstallStatusLine: %v", err)
	}

	b, err := os.ReadFile(p + ".bak")
	if err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	if string(b) != original {
		t.Errorf("backup = %q, want the untouched original %q", b, original)
	}
}

func TestInstallRefusesToClobberAnExistingStatusLine(t *testing.T) {
	p := settingsPath(t)
	if err := os.WriteFile(p, []byte(`{"statusLine":{"type":"command","command":"mine.sh"}}`), 0o600); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	err := InstallStatusLine(p, "claude-statusbar collect", false)
	if err == nil {
		t.Fatal("InstallStatusLine overwrote an existing statusLine, want a refusal")
	}
	if !strings.Contains(err.Error(), "passthrough") {
		t.Errorf("error %q should point at --passthrough", err)
	}
	if got := read(t, p)["statusLine"].(map[string]any)["command"]; got != "mine.sh" {
		t.Errorf("command = %v, want the original untouched", got)
	}
}

func TestInstallOverwritesWhenForced(t *testing.T) {
	p := settingsPath(t)
	if err := os.WriteFile(p, []byte(`{"statusLine":{"type":"command","command":"mine.sh"}}`), 0o600); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	if err := InstallStatusLine(p, "claude-statusbar collect", true); err != nil {
		t.Fatalf("InstallStatusLine(force): %v", err)
	}
	if got := read(t, p)["statusLine"].(map[string]any)["command"]; got != "claude-statusbar collect" {
		t.Errorf("command = %v, want it replaced", got)
	}
}

func TestInstallRejectsUnparseableSettingsInsteadOfDestroyingThem(t *testing.T) {
	p := settingsPath(t)
	if err := os.WriteFile(p, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	if err := InstallStatusLine(p, "claude-statusbar collect", false); err == nil {
		t.Fatal("InstallStatusLine accepted corrupt settings, want an error")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "{not json" {
		t.Errorf("settings = %q, want it left untouched", b)
	}
}

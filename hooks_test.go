package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallTrackingHookPreservesSettingsAndIsIdempotent(t *testing.T) {
	for _, tc := range []struct {
		provider string
		file     string
		matcher  string
	}{
		{"codex", filepath.Join(".codex", "hooks.json"), "startup|resume|compact|clear"},
		{"claude", filepath.Join(".claude", "settings.json"), "startup|resume|compact|clear|fork"},
	} {
		t.Run(tc.provider, func(t *testing.T) {
			base := t.TempDir()
			path := filepath.Join(base, tc.file)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			original := []byte(`{"model":"test-model","hooks":{"SessionStart":[{"matcher":"resume","hooks":[{"type":"command","command":"existing command"}]}],"Stop":[{"hooks":[{"type":"command","command":"other command"}]}]}}`)
			if err := os.WriteFile(path, original, 0600); err != nil {
				t.Fatal(err)
			}
			installed, err := installTrackingHook(tc.provider, base)
			if err != nil {
				t.Fatal(err)
			}
			if installed != path {
				t.Fatalf("path = %q, want %q", installed, path)
			}
			first, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var config struct {
				Model string `json:"model"`
				Hooks struct {
					SessionStart []struct {
						Matcher string `json:"matcher"`
						Hooks   []struct {
							Type    string `json:"type"`
							Command string `json:"command"`
						} `json:"hooks"`
					} `json:"SessionStart"`
					Stop json.RawMessage `json:"Stop"`
				} `json:"hooks"`
			}
			if err := json.Unmarshal(first, &config); err != nil {
				t.Fatal(err)
			}
			if config.Model != "test-model" || len(config.Hooks.Stop) == 0 {
				t.Fatalf("existing settings lost: %s", first)
			}
			if len(config.Hooks.SessionStart) != 2 || config.Hooks.SessionStart[0].Hooks[0].Command != "existing command" {
				t.Fatalf("existing hook lost: %s", first)
			}
			added := config.Hooks.SessionStart[1]
			if added.Matcher != tc.matcher || len(added.Hooks) != 1 || added.Hooks[0].Type != "command" || added.Hooks[0].Command != "tracking context" {
				t.Fatalf("unexpected Tracking hook: %s", first)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0600 {
				t.Fatalf("permissions = %o, want 600", info.Mode().Perm())
			}
			if _, err := installTrackingHook(tc.provider, base); err != nil {
				t.Fatal(err)
			}
			second, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(first, second) {
				t.Fatalf("second install changed settings:\n%s", second)
			}
		})
	}
}

func TestInstallTrackingHookRejectsInvalidConfigWithoutReplacingIt(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	invalid := []byte(`{"hooks":{"SessionStart":{"not":"an array"}}}`)
	if err := os.WriteFile(path, invalid, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := installTrackingHook("claude", base); err == nil {
		t.Fatal("expected malformed config error")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, invalid) {
		t.Fatalf("invalid config was modified: %s", data)
	}
}

func TestInstallTrackingHookRejectsUnknownProvider(t *testing.T) {
	if _, err := installTrackingHook("other", t.TempDir()); err == nil {
		t.Fatal("expected unsupported provider error")
	}
}

func TestInstallTrackingHookCreatesConfig(t *testing.T) {
	base := t.TempDir()
	path, err := installTrackingHook("codex", base)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(base, ".codex", "hooks.json") {
		t.Fatalf("unexpected path: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`"command": "tracking context"`)) {
		t.Fatalf("Tracking hook missing: %s", data)
	}
}

func TestInstallTrackingHookPreservesConfigSymlink(t *testing.T) {
	base := t.TempDir()
	configDir := filepath.Join(base, ".claude")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(base, "shared-settings.json")
	if err := os.WriteFile(linked, []byte(`{"custom":true}`), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(configDir, "settings.json")
	if err := os.Symlink(linked, path); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := installTrackingHook("claude", base); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("config symlink was replaced")
	}
	data, err := os.ReadFile(linked)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`"custom": true`)) || !bytes.Contains(data, []byte(`"command": "tracking context"`)) {
		t.Fatalf("linked config was not preserved and updated: %s", data)
	}
}

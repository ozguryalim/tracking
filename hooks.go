package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// installTrackingHook adds Tracking's optional session context hook to an
// existing Codex or Claude configuration without replacing unrelated settings.
func installTrackingHook(provider, base string) (string, error) {
	var target string
	var matcher string
	switch provider {
	case "codex":
		target = filepath.Join(base, ".codex", "hooks.json")
		matcher = "startup|resume|compact|clear"
	case "claude":
		target = filepath.Join(base, ".claude", "settings.json")
		matcher = "startup|resume|compact|clear|fork"
	default:
		return "", fmt.Errorf("unsupported hook provider %q", provider)
	}
	writeTarget := target
	if info, err := os.Lstat(target); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			writeTarget, err = filepath.EvalSymlinks(target)
			if err != nil {
				return "", fmt.Errorf("resolve %s: %w", target, err)
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	config := make(map[string]json.RawMessage)
	data, err := os.ReadFile(target)
	if err == nil {
		if err := decodeJSONObject(data, &config); err != nil {
			return "", fmt.Errorf("read %s: %w", target, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	hooks := make(map[string]json.RawMessage)
	if raw, ok := config["hooks"]; ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		if err := decodeJSONObject(raw, &hooks); err != nil {
			return "", fmt.Errorf("read hooks in %s: %w", target, err)
		}
	}

	var groups []json.RawMessage
	if raw, ok := hooks["SessionStart"]; ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		if err := json.Unmarshal(raw, &groups); err != nil {
			return "", fmt.Errorf("read SessionStart hooks in %s: %w", target, err)
		}
		if groups == nil {
			return "", fmt.Errorf("SessionStart hooks in %s must be an array", target)
		}
	}

	for _, rawGroup := range groups {
		var group map[string]json.RawMessage
		if err := decodeJSONObject(rawGroup, &group); err != nil {
			return "", fmt.Errorf("read SessionStart group in %s: %w", target, err)
		}
		var handlers []json.RawMessage
		if raw, ok := group["hooks"]; ok {
			if err := json.Unmarshal(raw, &handlers); err != nil || handlers == nil {
				return "", fmt.Errorf("read SessionStart handlers in %s: expected an array", target)
			}
		}
		for _, rawHandler := range handlers {
			var handler map[string]json.RawMessage
			if err := decodeJSONObject(rawHandler, &handler); err != nil {
				return "", fmt.Errorf("read SessionStart handler in %s: %w", target, err)
			}
			var kind, command string
			_ = json.Unmarshal(handler["type"], &kind)
			_ = json.Unmarshal(handler["command"], &command)
			if kind == "command" && command == "tracking context" {
				return target, nil
			}
		}
	}

	entry, err := json.Marshal(map[string]any{
		"matcher": matcher,
		"hooks":   []map[string]string{{"type": "command", "command": "tracking context"}},
	})
	if err != nil {
		return "", err
	}
	groups = append(groups, entry)
	if hooks["SessionStart"], err = json.Marshal(groups); err != nil {
		return "", err
	}
	if config["hooks"], err = json.Marshal(hooks); err != nil {
		return "", err
	}
	updated, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	updated = append(updated, '\n')
	if err := atomicWriteHookConfig(writeTarget, updated); err != nil {
		return "", err
	}
	return target, nil
}

func decodeJSONObject(data []byte, dst *map[string]json.RawMessage) error {
	if len(bytes.TrimSpace(data)) == 0 || bytes.TrimSpace(data)[0] != '{' {
		return fmt.Errorf("expected a JSON object")
	}
	return json.Unmarshal(data, dst)
}

func atomicWriteHookConfig(target string, data []byte) error {
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	mode := os.FileMode(0644)
	if info, err := os.Stat(target); err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.CreateTemp(dir, ".tracking-hook-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Chmod(mode); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), target)
}

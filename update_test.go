package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type updateRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn updateRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func updateTestResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestNewestStableReleaseUsesSemverNotReleaseOrder(t *testing.T) {
	releases := []trackingRelease{
		{TagName: "v0.1.0"},
		{TagName: "v0.1.1"},
		{TagName: "v9.0.0-rc.1", Prerelease: true},
		{TagName: "v8.0.0", Draft: true},
		{TagName: "v01.2.3"},
		{TagName: "v0.2.0"},
	}
	latest, found := newestStableRelease(releases)
	if !found || latest.TagName != "v0.2.0" {
		t.Fatalf("latest release = %+v, found = %v", latest, found)
	}
	for _, invalid := range []string{"dev", "v1.2", "v1.2.3-beta", "v1.02.3", "v1.2.3+build", "v1.2.-3"} {
		if _, ok := parseStableVersion(invalid); ok {
			t.Errorf("accepted unstable version %q", invalid)
		}
	}
}

func TestMaybeNotifyUpdateThrottlesAndNotifiesOncePerVersion(t *testing.T) {
	oldVersion := version
	oldNow := updateNow
	oldLocation := updateCacheLocation
	oldTerminal := updateOutputIsTerminal
	oldClient := updateCheckClient
	t.Cleanup(func() {
		version = oldVersion
		updateNow = oldNow
		updateCacheLocation = oldLocation
		updateOutputIsTerminal = oldTerminal
		updateCheckClient = oldClient
	})
	version = "v0.1.0"
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	updateNow = func() time.Time { return now }
	path := filepath.Join(t.TempDir(), "update.json")
	updateCacheLocation = func() (string, error) { return path, nil }
	updateOutputIsTerminal = func(io.Writer) bool { return true }
	latest, requests := "v0.1.1", 0
	updateCheckClient = &http.Client{Transport: updateRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		return updateTestResponse(fmt.Sprintf(`[{"tag_name":%q}]`, latest)), nil
	})}

	var output bytes.Buffer
	maybeNotifyUpdate(&output)
	if requests != 1 || !strings.Contains(output.String(), "v0.1.1") {
		t.Fatalf("first check: requests=%d output=%q", requests, output.String())
	}
	output.Reset()
	maybeNotifyUpdate(&output)
	if requests != 1 || output.Len() != 0 {
		t.Fatalf("cached check: requests=%d output=%q", requests, output.String())
	}
	now = now.Add(31 * time.Minute)
	maybeNotifyUpdate(&output)
	if requests != 2 || output.Len() != 0 {
		t.Fatalf("same release after expiry: requests=%d output=%q", requests, output.String())
	}
	latest = "v0.1.2"
	now = now.Add(31 * time.Minute)
	maybeNotifyUpdate(&output)
	if requests != 3 || !strings.Contains(output.String(), "v0.1.2") {
		t.Fatalf("new release: requests=%d output=%q", requests, output.String())
	}
	output.Reset()
	version = "dev"
	maybeNotifyUpdate(&output)
	if requests != 3 || output.Len() != 0 {
		t.Fatalf("development build checked implicitly: requests=%d output=%q", requests, output.String())
	}
}

func TestUpdateCheckReportsNewReleaseAndPlainUpdateDoesNotInstallNonInteractive(t *testing.T) {
	oldVersion := version
	oldClient := updateCheckClient
	oldTerminal := updateInputIsTerminal
	t.Cleanup(func() {
		version = oldVersion
		updateCheckClient = oldClient
		updateInputIsTerminal = oldTerminal
	})
	version = "v0.1.0"
	updateInputIsTerminal = func() bool { return false }
	requests := 0
	updateCheckClient = &http.Client{Transport: updateRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		return updateTestResponse(`[{"tag_name":"v0.1.1"},{"tag_name":"v0.1.0"}]`), nil
	})}
	var output bytes.Buffer
	if err := runUpdate([]string{"check"}, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "v0.1.0 -> v0.1.1") {
		t.Fatalf("check output = %q", output.String())
	}
	output.Reset()
	if err := runUpdate(nil, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && !strings.Contains(output.String(), "tracking update --yes") {
		t.Fatalf("noninteractive update output = %q", output.String())
	}
	if requests != 2 {
		t.Fatalf("expected a fresh check for each explicit command, got %d", requests)
	}
}

func TestUpdatePromptAndDevelopmentBuildGuard(t *testing.T) {
	oldVersion := version
	oldClient := updateCheckClient
	oldTerminal := updateInputIsTerminal
	oldInput := updatePromptInput
	t.Cleanup(func() {
		version = oldVersion
		updateCheckClient = oldClient
		updateInputIsTerminal = oldTerminal
		updatePromptInput = oldInput
	})
	updateCheckClient = &http.Client{Transport: updateRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return updateTestResponse(`[{"tag_name":"v0.1.1"}]`), nil
	})}
	version = "v0.1.0"
	updateInputIsTerminal = func() bool { return true }
	updatePromptInput = strings.NewReader("n\n")
	var output, prompt bytes.Buffer
	if runtime.GOOS != "windows" {
		if err := runUpdate(nil, &output, &prompt); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(prompt.String(), "[y/N]") || !strings.Contains(output.String(), "canceled") {
			t.Fatalf("prompt=%q output=%q", prompt.String(), output.String())
		}
	}
	version = "dev"
	if err := runUpdate([]string{"--yes"}, io.Discard, io.Discard); err == nil || !strings.Contains(err.Error(), "development build") {
		t.Fatalf("development build auto-update was allowed: %v", err)
	}
}

func TestMaybeNotifyUpdateSilentlyThrottlesOfflineChecks(t *testing.T) {
	oldVersion := version
	oldNow := updateNow
	oldLocation := updateCacheLocation
	oldTerminal := updateOutputIsTerminal
	oldClient := updateCheckClient
	t.Cleanup(func() {
		version = oldVersion
		updateNow = oldNow
		updateCacheLocation = oldLocation
		updateOutputIsTerminal = oldTerminal
		updateCheckClient = oldClient
	})
	version = "v0.1.0"
	updateNow = func() time.Time { return time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC) }
	path := filepath.Join(t.TempDir(), "update.json")
	updateCacheLocation = func() (string, error) { return path, nil }
	if err := writeUpdateCache(path, updateCache{
		LastChecked:   updateNow().Add(-31 * time.Minute),
		LatestVersion: "v0.1.1",
	}); err != nil {
		t.Fatal(err)
	}
	updateOutputIsTerminal = func(io.Writer) bool { return true }
	requests := 0
	updateCheckClient = &http.Client{Transport: updateRoundTripFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return nil, errors.New("offline")
	})}
	var output bytes.Buffer
	maybeNotifyUpdate(&output)
	maybeNotifyUpdate(&output)
	if requests != 1 || output.Len() != 0 {
		t.Fatalf("offline checks: requests=%d output=%q", requests, output.String())
	}
}

func TestInstallReleaseVerifiesChecksumBeforeReplacingExecutable(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("self-update is supported on macOS and Linux")
	}
	oldExecutable := updateExecutablePath
	oldDownloadClient := updateDownloadClient
	t.Cleanup(func() {
		updateExecutablePath = oldExecutable
		updateDownloadClient = oldDownloadClient
	})
	executable := filepath.Join(t.TempDir(), "tracking")
	if err := os.WriteFile(executable, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}
	updateExecutablePath = func() (string, error) { return executable, nil }
	name := "tracking_" + runtime.GOOS + "_" + runtime.GOARCH
	binary := "new binary"
	hash := sha256.Sum256([]byte(binary))
	checksum := hex.EncodeToString(hash[:])
	badChecksum := strings.Repeat("0", 64)
	assetURL := func(asset string) string {
		return "https://github.com/ozguryalim/tracking/releases/download/v0.1.1/" + asset
	}
	release := trackingRelease{TagName: "v0.1.1", Assets: []trackingAsset{
		{Name: name, BrowserDownloadURL: assetURL(name)},
		{Name: "checksums.txt", BrowserDownloadURL: assetURL("checksums.txt")},
	}}
	updateDownloadClient = &http.Client{Transport: updateRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/checksums.txt") {
			return updateTestResponse(badChecksum + "  " + name + "\n"), nil
		}
		return updateTestResponse(binary), nil
	})}
	if err := installRelease(release); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("bad checksum did not stop installation: %v", err)
	}
	data, err := os.ReadFile(executable)
	if err != nil || string(data) != "old binary" {
		t.Fatalf("failed installation changed executable: %q, %v", data, err)
	}
	updateDownloadClient = &http.Client{Transport: updateRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/checksums.txt") {
			return updateTestResponse(checksum + "  " + name + "\n"), nil
		}
		return updateTestResponse(binary), nil
	})}
	if err := installRelease(release); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(executable)
	if err != nil || string(data) != binary {
		t.Fatalf("verified installation failed: %q, %v", data, err)
	}
}

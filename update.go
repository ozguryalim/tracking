package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var version = "dev"

const (
	updateReleasesURL   = "https://api.github.com/repos/ozguryalim/tracking/releases?per_page=100"
	updateReleasePage   = "https://github.com/ozguryalim/tracking/releases"
	updateCheckInterval = 30 * time.Minute
	updateCheckTimeout  = 3 * time.Second
	updateMaxJSONBytes  = 4 << 20
	updateMaxAssetBytes = 150 << 20
)

var (
	updateNow                        = time.Now
	updateExecutablePath             = os.Executable
	updatePromptInput      io.Reader = os.Stdin
	updateInputIsTerminal            = func() bool { return isInteractiveFile(os.Stdin) }
	updateOutputIsTerminal           = func(out io.Writer) bool {
		file, ok := out.(*os.File)
		return ok && isInteractiveFile(file)
	}
	updateCacheLocation = func() (string, error) {
		dir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "tracking", "update.json"), nil
	}
	updateCheckClient    = &http.Client{Timeout: updateCheckTimeout}
	updateDownloadClient = &http.Client{Timeout: 2 * time.Minute}
)

type trackingRelease struct {
	TagName    string          `json:"tag_name"`
	Draft      bool            `json:"draft"`
	Prerelease bool            `json:"prerelease"`
	Assets     []trackingAsset `json:"assets"`
}

type trackingAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type updateCache struct {
	LastChecked     time.Time `json:"last_checked"`
	LatestVersion   string    `json:"latest_version"`
	NotifiedVersion string    `json:"notified_version"`
}

type stableVersion struct {
	major uint64
	minor uint64
	patch uint64
}

func parseStableVersion(value string) (stableVersion, bool) {
	value = strings.TrimPrefix(value, "v")
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return stableVersion{}, false
	}
	var numbers [3]uint64
	for index, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return stableVersion{}, false
		}
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return stableVersion{}, false
			}
		}
		number, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return stableVersion{}, false
		}
		numbers[index] = number
	}
	return stableVersion{numbers[0], numbers[1], numbers[2]}, true
}

func (value stableVersion) compare(other stableVersion) int {
	if value.major != other.major {
		if value.major < other.major {
			return -1
		}
		return 1
	}
	if value.minor != other.minor {
		if value.minor < other.minor {
			return -1
		}
		return 1
	}
	if value.patch < other.patch {
		return -1
	}
	if value.patch > other.patch {
		return 1
	}
	return 0
}

func newestStableRelease(releases []trackingRelease) (trackingRelease, bool) {
	var latest trackingRelease
	var latestNumber stableVersion
	found := false
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}
		number, valid := parseStableVersion(release.TagName)
		if valid && (!found || number.compare(latestNumber) > 0) {
			latest, latestNumber, found = release, number, true
		}
	}
	return latest, found
}

func fetchLatestRelease() (trackingRelease, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, updateReleasesURL, nil)
	if err != nil {
		return trackingRelease{}, false, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "tracking/"+version)
	response, err := updateCheckClient.Do(request)
	if err != nil {
		return trackingRelease{}, false, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return trackingRelease{}, false, fmt.Errorf("GitHub releases returned HTTP %d", response.StatusCode)
	}
	var releases []trackingRelease
	if err := json.NewDecoder(io.LimitReader(response.Body, updateMaxJSONBytes)).Decode(&releases); err != nil {
		return trackingRelease{}, false, fmt.Errorf("read GitHub releases: %w", err)
	}
	release, found := newestStableRelease(releases)
	return release, found, nil
}

func runVersion(args []string, out io.Writer) error {
	if len(args) != 0 {
		return fmt.Errorf("version does not accept arguments")
	}
	_, err := fmt.Fprintf(out, "tracking %s\n", version)
	return err
}

func isInteractiveFile(file *os.File) bool {
	if file == nil || file.Name() == os.DevNull {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func maybeNotifyUpdate(errOut io.Writer) {
	current, valid := parseStableVersion(version)
	if !valid || !updateOutputIsTerminal(errOut) {
		return
	}
	cachePath, err := updateCacheLocation()
	if err != nil {
		return
	}
	cache := readUpdateCache(cachePath)
	now := updateNow()
	if cache.LastChecked.IsZero() || now.Before(cache.LastChecked) || now.Sub(cache.LastChecked) >= updateCheckInterval {
		cache.LastChecked = now
		release, found, err := fetchLatestRelease()
		cache.LatestVersion = ""
		if err == nil {
			if found {
				cache.LatestVersion = release.TagName
			}
		}
		if err := writeUpdateCache(cachePath, cache); err != nil {
			return
		}
		if err != nil {
			return
		}
	}
	latest, valid := parseStableVersion(cache.LatestVersion)
	if !valid || latest.compare(current) <= 0 || cache.NotifiedVersion == cache.LatestVersion {
		return
	}
	cache.NotifiedVersion = cache.LatestVersion
	if err := writeUpdateCache(cachePath, cache); err != nil {
		return
	}
	_, _ = fmt.Fprintf(errOut, "Tracking %s is available. Run tracking update for details.\n", cache.LatestVersion)
}

func readUpdateCache(path string) updateCache {
	data, err := os.ReadFile(path)
	if err != nil {
		return updateCache{}
	}
	var cache updateCache
	if json.Unmarshal(data, &cache) != nil {
		return updateCache{}
	}
	return cache
}

func writeUpdateCache(path string, cache updateCache) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".update-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := temp.Chmod(0600); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

func runUpdate(args []string, out, errOut io.Writer) error {
	checkOnly, confirmed := false, false
	switch {
	case len(args) == 0:
	case len(args) == 1 && args[0] == "check":
		checkOnly = true
	case len(args) == 1 && args[0] == "--yes":
		confirmed = true
	default:
		return fmt.Errorf("usage: tracking update [check|--yes]")
	}
	release, found, err := fetchLatestRelease()
	if err != nil {
		return err
	}
	if !found {
		_, err := fmt.Fprintln(out, "No stable Tracking release is available yet.")
		return err
	}
	current, valid := parseStableVersion(version)
	latest, _ := parseStableVersion(release.TagName)
	if checkOnly {
		if !valid {
			_, err = fmt.Fprintf(out, "Installed: %s; latest stable: %s (%s).\n", version, release.TagName, releaseURL(release.TagName))
		} else if latest.compare(current) > 0 {
			_, err = fmt.Fprintf(out, "Update available: %s -> %s (%s).\n", version, release.TagName, releaseURL(release.TagName))
		} else {
			_, err = fmt.Fprintf(out, "Tracking is up to date (%s).\n", version)
		}
		return err
	}
	if !valid {
		return fmt.Errorf("cannot auto-update a development build; install a release from %s", releaseURL(release.TagName))
	}
	if latest.compare(current) <= 0 {
		_, err = fmt.Fprintf(out, "Tracking is up to date (%s).\n", version)
		return err
	}
	if runtime.GOOS == "windows" {
		_, err = fmt.Fprintf(out, "Download Tracking %s manually from %s\n", release.TagName, releaseURL(release.TagName))
		return err
	}
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return fmt.Errorf("automatic updates are unsupported on %s; use %s", runtime.GOOS, releaseURL(release.TagName))
	}
	if !confirmed {
		if !updateInputIsTerminal() {
			_, err = fmt.Fprintf(out, "Tracking %s is available. Run tracking update --yes to install it.\n", release.TagName)
			return err
		}
		if _, err := fmt.Fprintf(errOut, "Install Tracking %s? [y/N] ", release.TagName); err != nil {
			return err
		}
		answer, err := bufio.NewReader(updatePromptInput).ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer != "y" && answer != "yes" {
			_, err = fmt.Fprintln(out, "Update canceled.")
			return err
		}
	}
	if err := installRelease(release); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Updated Tracking to %s. If the dashboard is already running, use tracking stop before reopening it.\n", release.TagName)
	return err
}

func releaseURL(tag string) string {
	return updateReleasePage + "/tag/" + url.PathEscape(tag)
}

func findReleaseAsset(release trackingRelease, name string) (trackingAsset, error) {
	var asset trackingAsset
	found := false
	for _, candidate := range release.Assets {
		if candidate.Name == name {
			if found {
				return trackingAsset{}, fmt.Errorf("duplicate release asset %s", name)
			}
			asset, found = candidate, true
		}
	}
	if !found {
		return trackingAsset{}, fmt.Errorf("release %s has no %s asset", release.TagName, name)
	}
	u, err := url.Parse(asset.BrowserDownloadURL)
	expectedPath := "/ozguryalim/tracking/releases/download/" + release.TagName + "/" + name
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != expectedPath {
		return trackingAsset{}, fmt.Errorf("invalid download URL for %s", name)
	}
	return asset, nil
}

func downloadReleaseAsset(asset trackingAsset, maxBytes int64, writer io.Writer) error {
	request, err := http.NewRequest(http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "tracking/"+version)
	response, err := updateDownloadClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", asset.Name, response.StatusCode)
	}
	count, err := io.Copy(writer, io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return err
	}
	if count > maxBytes {
		return fmt.Errorf("release asset %s exceeds size limit", asset.Name)
	}
	return nil
}

func expectedChecksum(data []byte, assetName string) (string, error) {
	var expected string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != assetName {
			continue
		}
		if expected != "" {
			return "", fmt.Errorf("duplicate checksum for %s", assetName)
		}
		decoded, err := hex.DecodeString(fields[0])
		if err != nil || len(decoded) != sha256.Size {
			return "", fmt.Errorf("invalid checksum for %s", assetName)
		}
		expected = strings.ToLower(fields[0])
	}
	if expected == "" {
		return "", fmt.Errorf("no checksum for %s", assetName)
	}
	return expected, nil
}

func installRelease(release trackingRelease) error {
	assetName := "tracking_" + runtime.GOOS + "_" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}
	binaryAsset, err := findReleaseAsset(release, assetName)
	if err != nil {
		return err
	}
	checksumsAsset, err := findReleaseAsset(release, "checksums.txt")
	if err != nil {
		return err
	}
	var checksumData strings.Builder
	if err := downloadReleaseAsset(checksumsAsset, 1<<20, &checksumData); err != nil {
		return err
	}
	expected, err := expectedChecksum([]byte(checksumData.String()), assetName)
	if err != nil {
		return err
	}
	executable, err := updateExecutablePath()
	if err != nil {
		return err
	}
	info, err := os.Lstat(executable)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("installed executable is not a regular file: %s", executable)
	}
	temp, err := os.CreateTemp(filepath.Dir(executable), ".tracking-update-*")
	if err != nil {
		return fmt.Errorf("cannot write beside %s: %w", executable, err)
	}
	defer os.Remove(temp.Name())
	hash := sha256.New()
	if err := downloadReleaseAsset(binaryAsset, updateMaxAssetBytes, io.MultiWriter(temp, hash)); err != nil {
		temp.Close()
		return err
	}
	if hex.EncodeToString(hash.Sum(nil)) != expected {
		temp.Close()
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(temp.Name(), executable); err != nil {
		return fmt.Errorf("replace %s: %w", executable, err)
	}
	return nil
}

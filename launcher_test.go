package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLauncherAddressRequiresFixedLoopbackPort(t *testing.T) {
	accepted := []struct {
		name string
		args []string
		want string
	}{
		{name: "default", want: defaultListenAddress},
		{name: "localhost", args: []string{"--listen", "localhost:9001"}, want: "127.0.0.1:9001"},
		{name: "IPv4", args: []string{"--listen", "127.0.0.2:4157"}, want: "127.0.0.2:4157"},
		{name: "IPv6", args: []string{"--listen", "[::1]:9002"}, want: "[::1]:9002"},
	}
	for _, test := range accepted {
		t.Run(test.name, func(t *testing.T) {
			address, err := launcherAddress("dashboard", test.args, io.Discard)
			if err != nil || address != test.want {
				t.Fatalf("launcherAddress(%v) = %q, %v; want %q", test.args, address, err, test.want)
			}
		})
	}
	rejected := []struct {
		name string
		args []string
	}{
		{name: "all interfaces", args: []string{"--listen", "0.0.0.0:4157"}},
		{name: "empty host", args: []string{"--listen", ":4157"}},
		{name: "remote IPv4", args: []string{"--listen", "192.0.2.1:4157"}},
		{name: "remote name", args: []string{"--listen", "example.invalid:4157"}},
		{name: "remote IPv6", args: []string{"--listen", "[2001:db8::1]:4157"}},
		{name: "ephemeral port", args: []string{"--listen", "127.0.0.1:0"}},
		{name: "invalid port", args: []string{"--listen", "127.0.0.1:65536"}},
		{name: "missing port", args: []string{"--listen", "127.0.0.1"}},
		{name: "positional argument", args: []string{"extra"}},
	}
	for _, test := range rejected {
		t.Run(test.name, func(t *testing.T) {
			if address, err := launcherAddress("dashboard", test.args, io.Discard); err == nil {
				t.Fatalf("launcherAddress(%v) accepted %q", test.args, address)
			}
		})
	}
}

func TestHealthEndpointReportsManagedIdentity(t *testing.T) {
	store, _ := testStore(t)
	for _, managed := range []bool{false, true} {
		response := httptest.NewRecorder()
		newHTTPHandlerWithRuntime(store, managed).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("health status = %d; body=%s", response.Code, response.Body.String())
		}
		if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
			t.Fatalf("health content type = %q", contentType)
		}
		var health serverHealth
		if err := json.Unmarshal(response.Body.Bytes(), &health); err != nil {
			t.Fatal(err)
		}
		if !health.OK || health.App != "tracking" || health.DBKey != store.key || health.Managed != managed || health.PID != os.Getpid() {
			t.Fatalf("unexpected health identity: %+v", health)
		}
	}
}

func TestProbeServerDistinguishesTrackingDatabaseAndOtherServices(t *testing.T) {
	store, _ := testStore(t)
	trackingServer := httptest.NewServer(newHTTPHandlerWithRuntime(store, true))
	defer trackingServer.Close()
	address := strings.TrimPrefix(trackingServer.URL, "http://")

	health, running, err := probeServer(address, store.key)
	if err != nil || !running || !health.OK || health.App != "tracking" || health.DBKey != store.key || !health.Managed || health.PID != os.Getpid() {
		t.Fatalf("matching Tracking server: health=%+v running=%v err=%v", health, running, err)
	}
	_, running, err = probeServer(address, "different-database")
	if !running || err == nil || !strings.Contains(err.Error(), "another database") {
		t.Fatalf("database mismatch: running=%v err=%v", running, err)
	}

	unknownServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true,"app":"another-app","db_key":"`+store.key+`"}`)
	}))
	defer unknownServer.Close()
	_, running, err = probeServer(strings.TrimPrefix(unknownServer.URL, "http://"), store.key)
	if !running || err == nil || !strings.Contains(err.Error(), "different or older service") {
		t.Fatalf("unknown service: running=%v err=%v", running, err)
	}

	notFoundServer := httptest.NewServer(http.NotFoundHandler())
	defer notFoundServer.Close()
	_, running, err = probeServer(strings.TrimPrefix(notFoundServer.URL, "http://"), store.key)
	if !running || err == nil || !strings.Contains(err.Error(), "different service") {
		t.Fatalf("service without health endpoint: running=%v err=%v", running, err)
	}
}

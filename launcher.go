package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const defaultListenAddress = "127.0.0.1:4157"

type serverHealth struct {
	OK      bool   `json:"ok"`
	App     string `json:"app"`
	DBKey   string `json:"db_key"`
	Managed bool   `json:"managed"`
	PID     int    `json:"pid"`
}

func launcherAddress(command string, args []string, errOut io.Writer) (string, error) {
	flags := newFlags(command, errOut)
	listen := flags.String("listen", defaultListenAddress, "local dashboard address")
	if err := flags.Parse(args); err != nil {
		return "", err
	}
	if flags.NArg() != 0 {
		return "", fmt.Errorf("%s does not accept positional arguments", command)
	}
	host, port, err := net.SplitHostPort(*listen)
	if err != nil {
		return "", fmt.Errorf("invalid listen address: %w", err)
	}
	if host == "localhost" {
		host = "127.0.0.1"
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return "", fmt.Errorf("dashboard launcher requires a loopback address; use tracking serve for remote access")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", fmt.Errorf("dashboard launcher requires a fixed port from 1 to 65535")
	}
	return net.JoinHostPort(host, port), nil
}

func probeServer(address, key string) (serverHealth, bool, error) {
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	response, err := client.Get("http://" + address + "/api/health")
	if err != nil {
		return serverHealth{}, false, nil
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return serverHealth{}, true, fmt.Errorf("port %s is occupied by a different service", address)
	}
	var health serverHealth
	if err := json.NewDecoder(io.LimitReader(response.Body, 4096)).Decode(&health); err != nil || !health.OK || health.App != "tracking" {
		return serverHealth{}, true, fmt.Errorf("port %s is occupied by a different or older service", address)
	}
	if health.DBKey != key {
		return serverHealth{}, true, fmt.Errorf("Tracking on %s uses another database; choose a different --listen port", address)
	}
	return health, true, nil
}

func runDashboard(args []string, out, errOut io.Writer) error {
	address, err := launcherAddress("dashboard", args, errOut)
	if err != nil {
		return err
	}
	path, err := databasePath()
	if err != nil {
		return err
	}
	key := databaseKey(path)
	_, running, err := probeServer(address, key)
	if err != nil {
		return err
	}
	if !running {
		if err := startBackgroundServer(address); err != nil {
			return err
		}
		deadline := time.Now().Add(5 * time.Second)
		for {
			_, running, err = probeServer(address, key)
			if err != nil {
				return err
			}
			if running {
				break
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("dashboard did not start on %s; check the Tracking server log", address)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	url := "http://" + address
	fmt.Fprintf(out, "Tracking dashboard: %s\n", url)
	if err := openBrowser(url); err != nil {
		fmt.Fprintf(errOut, "Could not open a browser automatically: %v\n", err)
	}
	return nil
}

func startBackgroundServer(address string) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	logDir := filepath.Join(configDir, "tracking")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return err
	}
	logPath := filepath.Join(logDir, "server.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	command := exec.Command(executable, "serve", "--listen", address, "--managed")
	command.SysProcAttr = detachedProcessAttributes()
	command.Stdout = logFile
	command.Stderr = logFile
	if err := command.Start(); err != nil {
		return err
	}
	if err := command.Process.Release(); err != nil {
		return err
	}
	return nil
}

func runStop(args []string, out, errOut io.Writer) error {
	address, err := launcherAddress("stop", args, errOut)
	if err != nil {
		return err
	}
	path, err := databasePath()
	if err != nil {
		return err
	}
	health, running, err := probeServer(address, databaseKey(path))
	if err != nil {
		return err
	}
	if !running {
		fmt.Fprintln(out, "Tracking dashboard is not running.")
		return nil
	}
	if !health.Managed || health.PID <= 0 {
		return fmt.Errorf("Tracking on %s runs in the foreground; stop it in its terminal", address)
	}
	process, err := os.FindProcess(health.PID)
	if err != nil {
		return err
	}
	if err := process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_, running, _ = probeServer(address, databaseKey(path))
		if !running {
			fmt.Fprintln(out, "Tracking dashboard stopped.")
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("dashboard on %s did not stop", address)
}

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

const (
	launchdLabel = "com.mehdir.daemon"
	systemdUnit  = "mehdir.service"
)

// launchd plist template
var launchdTmpl = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.ExePath}}</string>
        <string>daemon</string>
        <string>run</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
</dict>
</plist>
`))

// systemd unit template
var systemdTmpl = template.Must(template.New("unit").Parse(`[Unit]
Description=mehdir temporary directory cleanup daemon

[Service]
ExecStart={{.ExePath}} daemon run
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`))

func launchdPlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", launchdLabel+".plist")
}

func systemdUnitPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user", systemdUnit)
}

func logPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "mehdir")
	os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "daemon.log")
}

// Install writes the service file and starts the daemon.
func Install() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable: %w", err)
	}
	// Resolve symlinks to get the real path.
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return installLaunchd(exe)
	case "linux":
		return installSystemd(exe)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func installLaunchd(exe string) error {
	path := launchdPlistPath()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating plist: %w", err)
	}
	defer f.Close()

	data := struct {
		Label   string
		ExePath string
		LogPath string
	}{launchdLabel, exe, logPath()}

	if err := launchdTmpl.Execute(f, data); err != nil {
		return fmt.Errorf("writing plist: %w", err)
	}

	out, err := exec.Command("launchctl", "load", "-w", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl load: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func installSystemd(exe string) error {
	path := systemdUnitPath()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating unit file: %w", err)
	}
	defer f.Close()

	data := struct{ ExePath string }{exe}
	if err := systemdTmpl.Execute(f, data); err != nil {
		return fmt.Errorf("writing unit file: %w", err)
	}

	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("daemon-reload: %s: %w", strings.TrimSpace(string(out)), err)
	}
	if out, err := exec.Command("systemctl", "--user", "enable", "--now", systemdUnit).CombinedOutput(); err != nil {
		return fmt.Errorf("enable: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Uninstall stops the daemon and removes the service file.
func Uninstall() error {
	switch runtime.GOOS {
	case "darwin":
		return uninstallLaunchd()
	case "linux":
		return uninstallSystemd()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func uninstallLaunchd() error {
	path := launchdPlistPath()
	// Unload (ignore errors if not loaded).
	exec.Command("launchctl", "unload", "-w", path).Run()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing plist: %w", err)
	}
	return nil
}

func uninstallSystemd() error {
	exec.Command("systemctl", "--user", "disable", "--now", systemdUnit).Run()
	path := systemdUnitPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing unit file: %w", err)
	}
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	return nil
}

// Start starts the service.
func Start() error {
	switch runtime.GOOS {
	case "darwin":
		path := launchdPlistPath()
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("daemon not installed (run: mehdir daemon install)")
		}
		out, err := exec.Command("launchctl", "load", "-w", path).CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl load: %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	case "linux":
		out, err := exec.Command("systemctl", "--user", "start", systemdUnit).CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl start: %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Stop stops the service.
func Stop() error {
	switch runtime.GOOS {
	case "darwin":
		path := launchdPlistPath()
		out, err := exec.Command("launchctl", "unload", path).CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl unload: %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	case "linux":
		out, err := exec.Command("systemctl", "--user", "stop", systemdUnit).CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl stop: %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Status returns a human-readable status string.
func Status() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("launchctl", "list", launchdLabel).CombinedOutput()
		if err != nil {
			return "not running", nil
		}
		return fmt.Sprintf("running\n%s", strings.TrimSpace(string(out))), nil
	case "linux":
		out, _ := exec.Command("systemctl", "--user", "is-active", systemdUnit).CombinedOutput()
		status := strings.TrimSpace(string(out))
		if status == "active" {
			return "running", nil
		}
		return status, nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Installed returns true if the service file exists.
func Installed() bool {
	switch runtime.GOOS {
	case "darwin":
		_, err := os.Stat(launchdPlistPath())
		return err == nil
	case "linux":
		_, err := os.Stat(systemdUnitPath())
		return err == nil
	default:
		return false
	}
}

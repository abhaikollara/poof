# mehdir

Temporary directories that disappear after their TTL expires.

## Install

### Homebrew

```sh
brew install abhaikollara/tap/mehdir
brew services start mehdir  # start the cleanup daemon
```

### From source

```sh
go install abhai.dev/mehdir/cmd/mehdir@latest
mehdir daemon install  # install the cleanup daemon
```

### Build locally

```sh
make build  # binary in bin/
```

## Usage

### Create a temp directory

```sh
# Auto-generated name, default 1 hour TTL
mehdir

# Auto-generated name, 5 minute TTL
mehdir 5m

# Named directory, 2 hour TTL
mehdir new myproject 2h

# Absolute path
mehdir new /tmp/scratch 30m

# Shell-friendly: cd into it
cd "$(mehdir new workspace 1d)"
```

TTL supports `s`, `m`, `h`, `d` (days), and `w` (weeks): `30m`, `2h`, `1d12h`, `2w`.

### Track an existing directory

```sh
mehdir add ./my-existing-dir 2h
```

### List active directories

```sh
mehdir ls
mehdir ls --json
```

### Extend the TTL

```sh
# Reset expiry to 3 hours from now (replaces, not additive)
mehdir extend /path/to/myproject 3h
```

### Remove immediately

```sh
mehdir rm /path/to/myproject
```

### Delete everything

```sh
mehdir cleanslate  # asks for confirmation
```

### Force cleanup of expired entries

```sh
mehdir clean
```

### Remove orphaned registry entries

```sh
mehdir gc
```

### Configuration

```sh
mehdir config               # show current config
mehdir config prefix        # show auto-generated directory prefix
mehdir config prefix tmp-   # set prefix to "tmp-" (default: "mehdir-")
mehdir config ttl           # show default TTL
mehdir config ttl 30m       # set default TTL to 30 minutes (default: "1h")
```

### Daemon

A background daemon automatically deletes expired directories. It auto-installs as a system service on first `mehdir new`.

```sh
mehdir daemon install   # install and start as a system service
mehdir daemon uninstall # stop and remove the service
mehdir daemon start     # start the service
mehdir daemon stop      # stop the service
mehdir daemon status    # check if it's running
```

On macOS this uses **launchd** (`~/Library/LaunchAgents/com.mehdir.daemon.plist`).
On Linux this uses **systemd** (`~/.config/systemd/user/mehdir.service`).

The daemon polls every 10 seconds. The service auto-restarts on crash and starts on login. Lazy sweep on each command still runs as a fallback.

Logs are at `~/.config/mehdir/daemon.log`.

## How it works

- `mehdir new mydir 1h` creates `mydir` directly and tracks it. Without a name, it creates a prefixed directory in the current directory.
- `mehdir add` tracks an existing directory without creating it.
- A background daemon (launchd/systemd) polls every 10s and deletes expired directories. It auto-installs on first `mehdir new`.
- Every command also runs a lazy sweep as a fallback in case the daemon isn't running.
- The registry is written atomically (write to `.tmp`, then rename) and protected by a file lock for concurrent access.

## Safety guardrails

Before deleting any directory, mehdir verifies:

1. The path is absolute.
2. The path is under a known allowed prefix (`$TMPDIR`, `/tmp`, `/var/tmp`, or any parent directory previously used with `mehdir new`).
3. The path is not `/`, `/tmp`, `/var/tmp`, or the user's home directory.

If any check fails, the entry is skipped and an error is logged to stderr.

## Configuration

The registry lives at `$XDG_CONFIG_HOME/mehdir/registry.json` (default: `~/.config/mehdir/registry.json`). There is no other configuration file.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (bad flag, path not found) |
| 2 | Internal error (registry corrupt, lock timeout) |

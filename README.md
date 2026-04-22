# poof

Temporary directories that disappear after their TTL expires.

## Install

```sh
go install abhai.dev/poof/cmd/poof@latest
```

Or build from source:

```sh
make build
```

## Usage

### Create a temp directory

```sh
# Named directory in the current dir, 5 minute TTL
poof new myproject 5m

# Absolute path, 2 hour TTL
poof new /tmp/scratch 2h

# Auto-generated name (poof-XXXXXX) in current dir, 30 minute TTL
poof new 30m

# Auto-generated name, default 1 hour TTL
poof new

# Shell-friendly: cd into it
cd "$(poof new workspace 1d)"
```

TTL supports `s`, `m`, `h`, `d` (days), and `w` (weeks): `30m`, `2h`, `1d12h`, `2w`.

### List active directories

```sh
poof ls
poof ls --json
```

### Extend the TTL

```sh
# Reset expiry to 3 hours from now (replaces, not additive)
poof extend /path/to/myproject 3h
```

### Remove immediately

```sh
poof rm /path/to/myproject
```

### Force cleanup of expired entries

```sh
poof clean
```

### Remove orphaned registry entries

```sh
poof gc
```

### Daemon

A background daemon automatically deletes expired directories. It starts automatically on `poof new`, or you can manage it manually:

```sh
poof daemon start     # start the background daemon
poof daemon stop      # stop it
poof daemon status    # check if it's running
```

The daemon polls every 10 seconds. Lazy sweep on each command still runs as a fallback.

## How it works

- `poof new mydir 1h` creates `mydir` directly and tracks it. Without a name, it creates a `poof-XXXXXX` directory in the current directory.
- A background daemon polls every 10s and deletes expired directories. It auto-starts on `poof new`.
- Every command also runs a lazy sweep as a fallback in case the daemon isn't running.
- The registry is written atomically (write to `.tmp`, then rename) and protected by a file lock for concurrent access.

## Safety guardrails

Before deleting any directory, poof verifies:

1. The path is absolute.
2. The path is under a known allowed prefix (`$TMPDIR`, `/tmp`, `/var/tmp`, or any parent directory previously used with `poof new`).
3. The path is not `/`, `/tmp`, `/var/tmp`, or the user's home directory.

If any check fails, the entry is skipped and an error is logged to stderr.

## Configuration

The registry lives at `$XDG_CONFIG_HOME/poof/registry.json` (default: `~/.config/poof/registry.json`). There is no other configuration file.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (bad flag, path not found) |
| 2 | Internal error (registry corrupt, lock timeout) |

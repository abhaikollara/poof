# poof

Temporary directories that disappear after their TTL expires. No daemon, no cron — just lazy cleanup on every invocation.

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
# Default: 1 hour TTL
cd "$(poof new)"

# Custom TTL and label
poof new --ttl 2h --label scratch

# Supports d (days) and w (weeks)
poof new --ttl 7d --label weekly-build
```

### List active directories

```sh
poof ls
poof ls --json
```

### Extend the TTL

```sh
# Reset expiry to 3 hours from now
poof extend scratch --ttl 3h
```

### Remove immediately

```sh
poof rm scratch
poof rm /tmp/poof-a4f2b9
```

### Force cleanup of expired entries

```sh
poof clean
```

### Remove orphaned registry entries

```sh
poof gc
```

## How it works

- `poof new` creates a directory via `os.MkdirTemp` (pattern `poof-XXXXXX`, mode 0700) and registers it in `~/.config/poof/registry.json`.
- Every command runs a lazy sweep that removes directories whose TTL has expired.
- The registry is written atomically (write to `.tmp`, then rename) and protected by a file lock for concurrent access.

## Safety guardrails

Before deleting any directory, poof verifies:

1. The path is absolute.
2. The path is under `$TMPDIR`, `/tmp`, or `/var/tmp`.
3. The path is not `/`, `/tmp`, `/var/tmp`, or the user's home directory.
4. The basename starts with `poof-`.

If any check fails, the entry is skipped and an error is logged to stderr.

## Configuration

The registry lives at `$XDG_CONFIG_HOME/poof/registry.json` (default: `~/.config/poof/registry.json`). There is no other configuration file.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (bad flag, label not found) |
| 2 | Internal error (registry corrupt, lock timeout) |

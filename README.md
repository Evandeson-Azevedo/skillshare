<p align="center">
  <img src=".github/assets/logo.png" alt="skillshare" width="200">
</p>

<h1 align="center">skillshare</h1>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/runkids/skillshare" alt="Go Version"></a>
  <a href="https://github.com/runkids/skillshare/releases"><img src="https://img.shields.io/github/v/release/runkids/skillshare" alt="Release"></a>
</p>

<p align="center">
  <strong>Sync skills to all your AI CLI tools with one command</strong><br>
  Supports Claude Code, Codex CLI, Cursor, Gemini CLI, OpenCode
</p>

<p align="center">
  <img src=".github/assets/demo.gif" alt="skillshare demo" width="600">
</p>

<p align="center">
  <a href="#installation">Install</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#commands">Commands</a> •
  <a href="#faq">FAQ</a>
</p>

## Installation

```bash
brew install runkids/tap/skillshare
```

> Other methods: [Detailed Installation](#detailed-installation)

## Quick Start

```bash
skillshare init    # Auto-detects installed CLIs, sets up git
skillshare sync    # Syncs skills to all targets
```

Done! Your skills are now synced across all AI CLI tools.

## How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                  ~/.config/skillshare/skills/               │
│         my-skill/   another-skill/   shared-util/           │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
              ▼               ▼               ▼
       ┌───────────┐   ┌───────────┐   ┌───────────┐
       │  Claude   │   │   Codex   │   │  Gemini   │
       │  skills/  │   │  skills/  │   │  skills/  │
       └───────────┘   └───────────┘   └───────────┘
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize skillshare, detect CLIs, setup git |
| `sync` | Sync skills to all targets |
| `status` | Show source, targets, and sync state |
| `diff` | Show differences between source and targets |
| `pull` | Pull skills from target back to source |
| `backup` | Manually backup targets |
| `restore` | Restore from backup |
| `doctor` | Diagnose configuration issues |
| `target` | Manage targets (add/remove/list/mode) |

---

# Detailed Documentation

## Detailed Installation

### macOS

```bash
# Apple Silicon (M1/M2/M3/M4)
curl -sL https://github.com/runkids/skillshare/releases/latest/download/skillshare_darwin_arm64.tar.gz | tar xz
sudo mv skillshare /usr/local/bin/

# Intel
curl -sL https://github.com/runkids/skillshare/releases/latest/download/skillshare_darwin_amd64.tar.gz | tar xz
sudo mv skillshare /usr/local/bin/
```

### Linux

```bash
# x86_64
curl -sL https://github.com/runkids/skillshare/releases/latest/download/skillshare_linux_amd64.tar.gz | tar xz
sudo mv skillshare /usr/local/bin/

# ARM64
curl -sL https://github.com/runkids/skillshare/releases/latest/download/skillshare_linux_arm64.tar.gz | tar xz
sudo mv skillshare /usr/local/bin/
```

### Windows

Download from [Releases](https://github.com/runkids/skillshare/releases) and add to PATH.

### Uninstall

```bash
brew uninstall skillshare              # Homebrew
sudo rm /usr/local/bin/skillshare      # Manual install
rm -rf ~/.config/skillshare            # Config & data (optional)
```

## Sync Modes

| Mode | Behavior | When to Use |
|------|----------|-------------|
| `merge` | Each skill symlinked individually. Local skills preserved. | **Recommended.** Safe, flexible. |
| `symlink` | Entire directory becomes symlink. All targets identical. | When you want exact copies everywhere. |

Change mode:

```bash
skillshare target claude --mode merge
skillshare sync
```

### ⚠️ Symlink Safety

Deleting through a symlinked target **deletes the source**:

```bash
rm -rf ~/.codex/skills/my-skill  # ❌ Deletes from SOURCE!
skillshare target remove codex   # ✅ Safe way to unlink
```

## Backup & Restore

Backups are created **automatically** before `sync` and `target remove`.

Location: `~/.config/skillshare/backups/<timestamp>/`

```bash
skillshare backup              # Manual backup all targets
skillshare backup claude       # Backup specific target
skillshare backup --list       # List all backups
skillshare backup --cleanup    # Remove old backups

skillshare restore claude      # Restore from latest backup
skillshare restore claude --from 2026-01-14_21-22-18  # Specific backup
```

> **Note:** In `symlink` mode, backups are skipped (no local data to backup).

## Configuration

Config file: `~/.config/skillshare/config.yaml`

```yaml
source: ~/.config/skillshare/skills
mode: merge
targets:
  claude:
    path: ~/.claude/skills
  codex:
    path: ~/.codex/skills
    mode: symlink  # Override default mode
  cursor:
    path: ~/.cursor/skills
ignore:
  - "**/.DS_Store"
  - "**/.git/**"
```

### Managing Targets

```bash
skillshare target list                        # List all targets
skillshare target claude                      # Show target info
skillshare target claude --mode merge         # Change mode
skillshare target add myapp ~/.myapp/skills   # Add custom target
skillshare target remove myapp                # Remove target
```

## FAQ

**How do I sync across multiple machines?**

```bash
# Machine A: Push to remote
cd ~/.config/skillshare/skills
git remote add origin git@github.com:you/my-skills.git
git push -u origin main

# Machine B: Clone and init
git clone git@github.com:you/my-skills.git ~/.config/skillshare/skills
skillshare init --source ~/.config/skillshare/skills
skillshare sync
```

**What happens if I modify a skill in the target directory?**

Since targets are symlinks, changes are made directly to the source. All targets see the change immediately.

**How do I keep a CLI-specific skill?**

Use `merge` mode. Local skills in the target won't be overwritten.

**What if I accidentally delete a skill through a symlink?**

If you have git initialized (recommended), recover with:

```bash
cd ~/.config/skillshare/skills
git checkout -- deleted-skill/
```

## Contributing

```bash
git clone https://github.com/runkids/skillshare.git
cd skillshare
go build -o bin/skillshare ./cmd/skillshare
go test ./...
```

[Open an issue](https://github.com/runkids/skillshare/issues) for bugs or feature requests.

## License

MIT

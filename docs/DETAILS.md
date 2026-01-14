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

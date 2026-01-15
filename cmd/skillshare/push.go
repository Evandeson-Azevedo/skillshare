package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/ui"
)

func cmdPush(args []string) error {
	dryRun := false
	message := "Update skills"

	// Parse args
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dry-run", "-n":
			dryRun = true
		case "-m", "--message":
			if i+1 < len(args) {
				i++
				message = args[i]
			}
		default:
			if strings.HasPrefix(arg, "-m=") {
				message = strings.TrimPrefix(arg, "-m=")
			} else if strings.HasPrefix(arg, "--message=") {
				message = strings.TrimPrefix(arg, "--message=")
			}
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config not found: run 'skillshare init' first")
	}

	ui.Header("Pushing to remote")

	// Check if source is a git repo
	gitDir := cfg.Source + "/.git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		ui.Error("Source is not a git repository")
		ui.Info("  Run: cd %s && git init", cfg.Source)
		return nil
	}

	// Check if remote exists
	cmd := exec.Command("git", "remote")
	cmd.Dir = cfg.Source
	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		ui.Error("No git remote configured")
		ui.Info("  Run: cd %s && git remote add origin <url>", cfg.Source)
		ui.Info("  Or:  skillshare init --remote <url>")
		return nil
	}

	// Check for changes
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = cfg.Source
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	hasChanges := len(strings.TrimSpace(string(output))) > 0

	if dryRun {
		ui.Warning("[dry-run] No changes will be made")
		fmt.Println()
		if hasChanges {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			ui.Info("Would stage %d file(s):", len(lines))
			for _, line := range lines {
				ui.Info("  %s", line)
			}
			ui.Info("Would commit with message: %s", message)
		} else {
			ui.Info("No changes to commit")
		}
		ui.Info("Would push to remote")
		return nil
	}

	// Stage all changes
	if hasChanges {
		ui.Info("Staging changes...")
		cmd = exec.Command("git", "add", "-A")
		cmd.Dir = cfg.Source
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stage changes: %w", err)
		}

		// Commit
		ui.Info("Committing...")
		cmd = exec.Command("git", "commit", "-m", message)
		cmd.Dir = cfg.Source
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}
	} else {
		ui.Info("No changes to commit")
	}

	// Push
	ui.Info("Pushing to remote...")
	cmd = exec.Command("git", "push")
	cmd.Dir = cfg.Source
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println()
		ui.Error("Push failed - remote may have newer changes")
		ui.Info("  Run: skillshare pull --remote")
		ui.Info("  Then: skillshare push")
		return nil
	}

	fmt.Println()
	ui.Success("Pushed to remote")
	return nil
}

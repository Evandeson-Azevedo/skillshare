package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/install"
	"skillshare/internal/sync"
	"skillshare/internal/ui"
	"skillshare/internal/utils"
)

func cmdList(args []string) error {
	var verbose bool

	// Parse arguments
	for _, arg := range args {
		switch arg {
		case "--verbose", "-v":
			verbose = true
		case "--help", "-h":
			printListHelp()
			return nil
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown option: %s", arg)
			}
		}
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Discover all skills recursively
	discovered, err := sync.DiscoverSourceSkills(cfg.Source)
	if err != nil {
		return fmt.Errorf("cannot discover skills: %w", err)
	}

	// Get tracked repos
	trackedRepos, _ := install.GetTrackedRepos(cfg.Source)

	// Build skill entries with metadata
	var skills []skillEntry
	for _, d := range discovered {
		entry := skillEntry{
			Name:     d.FlatName,
			IsNested: d.IsInRepo || utils.HasNestedSeparator(d.FlatName),
		}

		// Determine repo name if in tracked repo
		if d.IsInRepo {
			parts := strings.SplitN(d.RelPath, "/", 2)
			if len(parts) > 0 {
				entry.RepoName = parts[0]
			}
		}

		// Read metadata if available
		if meta, err := install.ReadMeta(d.SourcePath); err == nil && meta != nil {
			entry.Source = meta.Source
			entry.Type = meta.Type
			entry.InstalledAt = meta.InstalledAt.Format("2006-01-02")
		}

		skills = append(skills, entry)
	}

	if len(skills) == 0 && len(trackedRepos) == 0 {
		ui.Info("No skills installed")
		ui.Info("Use 'skillshare install <source>' to install a skill")
		return nil
	}

	// Display skills
	if len(skills) > 0 {
		ui.Header("Installed skills")

		// Calculate max name length for alignment (compact mode only)
		maxNameLen := 0
		if !verbose {
			for _, s := range skills {
				if len(s.Name) > maxNameLen {
					maxNameLen = len(s.Name)
				}
			}
		}

		for _, s := range skills {
			if verbose {
				// Verbose mode: show full details
				fmt.Printf("  %s%s%s\n", ui.Cyan, s.Name, ui.Reset)
				if s.RepoName != "" {
					fmt.Printf("    %sTracked repo:%s %s\n", ui.Gray, ui.Reset, s.RepoName)
				}
				if s.Source != "" {
					fmt.Printf("    %sSource:%s      %s\n", ui.Gray, ui.Reset, s.Source)
					fmt.Printf("    %sType:%s        %s\n", ui.Gray, ui.Reset, s.Type)
					fmt.Printf("    %sInstalled:%s   %s\n", ui.Gray, ui.Reset, s.InstalledAt)
				} else {
					fmt.Printf("    %sSource:%s      (local - no metadata)\n", ui.Gray, ui.Reset)
				}
				fmt.Println()
			} else {
				// Compact mode: aligned skill name + source info
				var suffix string
				if s.RepoName != "" {
					suffix = fmt.Sprintf("tracked: %s", s.RepoName)
				} else if s.Source != "" {
					suffix = abbreviateSource(s.Source)
				} else {
					suffix = "local"
				}
				// Use dynamic width formatting with icon
				format := fmt.Sprintf("  %s→%s %%-%ds  %s%%s%s\n", ui.Cyan, ui.Reset, maxNameLen, ui.Gray, ui.Reset)
				fmt.Printf(format, s.Name, suffix)
			}
		}
	}

	// Display tracked repos section
	if len(trackedRepos) > 0 {
		fmt.Println()
		ui.Header("Tracked repositories")

		for _, repoName := range trackedRepos {
			repoPath := filepath.Join(cfg.Source, repoName)
			// Count skills in this repo
			skillCount := 0
			for _, d := range discovered {
				if d.IsInRepo && strings.HasPrefix(d.RelPath, repoName+"/") {
					skillCount++
				}
			}
			// Check git status and display with appropriate icon
			if isDirty, _ := isRepoDirty(repoPath); isDirty {
				ui.ListItem("warning", repoName, fmt.Sprintf("%d skills, has changes", skillCount))
			} else {
				ui.ListItem("success", repoName, fmt.Sprintf("%d skills, up-to-date", skillCount))
			}
		}
	}

	if !verbose && len(skills) > 0 {
		fmt.Println()
		ui.Info("Use --verbose for more details")
	}

	return nil
}

type skillEntry struct {
	Name        string
	Source      string
	Type        string
	InstalledAt string
	IsNested    bool
	RepoName    string
}

// abbreviateSource shortens long sources for display
func abbreviateSource(source string) string {
	// Remove https:// prefix
	source = strings.TrimPrefix(source, "https://")
	source = strings.TrimPrefix(source, "http://")

	// Truncate if too long
	if len(source) > 50 {
		return source[:47] + "..."
	}
	return source
}

func printListHelp() {
	fmt.Println(`Usage: skillshare list [options]

List all installed skills in the source directory.

Options:
  --verbose, -v   Show detailed information (source, type, install date)
  --help, -h      Show this help

Examples:
  skillshare list
  skillshare list --verbose`)
}

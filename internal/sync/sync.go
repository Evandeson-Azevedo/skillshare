package sync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/runkids/skillshare/internal/config"
)

// TargetStatus represents the state of a target
type TargetStatus int

const (
	StatusUnknown TargetStatus = iota
	StatusLinked               // Target is a symlink pointing to source
	StatusNotExist             // Target doesn't exist
	StatusHasFiles             // Target exists with files (needs migration)
	StatusConflict             // Target is a symlink pointing elsewhere
	StatusBroken               // Target is a broken symlink
)

func (s TargetStatus) String() string {
	switch s {
	case StatusLinked:
		return "linked"
	case StatusNotExist:
		return "not exist"
	case StatusHasFiles:
		return "has files"
	case StatusConflict:
		return "conflict"
	case StatusBroken:
		return "broken"
	default:
		return "unknown"
	}
}

// CheckStatus checks the status of a target
func CheckStatus(targetPath, sourcePath string) TargetStatus {
	info, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return StatusNotExist
		}
		return StatusUnknown
	}

	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(targetPath)
		if err != nil {
			return StatusUnknown
		}

		// Check if symlink points to our source
		absLink := link
		if !filepath.IsAbs(link) {
			absLink = filepath.Join(filepath.Dir(targetPath), link)
		}
		absSource, _ := filepath.Abs(sourcePath)
		absLink, _ = filepath.Abs(absLink)

		if absLink == absSource {
			// Verify the link is not broken
			if _, err := os.Stat(targetPath); err != nil {
				return StatusBroken
			}
			return StatusLinked
		}
		return StatusConflict
	}

	// It's a directory with files
	if info.IsDir() {
		return StatusHasFiles
	}

	return StatusUnknown
}

// MigrateToSource moves files from target to source, then creates symlink
func MigrateToSource(targetPath, sourcePath string) error {
	// Ensure source parent directory exists
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		return fmt.Errorf("failed to create source parent: %w", err)
	}

	// Check if source already exists
	if _, err := os.Stat(sourcePath); err == nil {
		// Source exists - merge files
		if err := mergeDirectories(targetPath, sourcePath); err != nil {
			return fmt.Errorf("failed to merge directories: %w", err)
		}
		// Remove original target
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to remove target after merge: %w", err)
		}
	} else {
		// Source doesn't exist - just move
		if err := os.Rename(targetPath, sourcePath); err != nil {
			// Cross-device? Try copy then delete
			if err := copyDirectory(targetPath, sourcePath); err != nil {
				return fmt.Errorf("failed to copy to source: %w", err)
			}
			if err := os.RemoveAll(targetPath); err != nil {
				return fmt.Errorf("failed to remove original after copy: %w", err)
			}
		}
	}

	return nil
}

// CreateSymlink creates a symlink from target to source
func CreateSymlink(targetPath, sourcePath string) error {
	// Ensure target parent exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create target parent: %w", err)
	}

	// Create symlink
	if err := os.Symlink(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// SyncTarget performs the sync operation for a single target
func SyncTarget(name string, target config.TargetConfig, sourcePath string, dryRun bool) error {
	status := CheckStatus(target.Path, sourcePath)

	switch status {
	case StatusLinked:
		// Already correct
		return nil

	case StatusNotExist:
		if dryRun {
			fmt.Printf("[dry-run] Would create symlink: %s -> %s\n", target.Path, sourcePath)
			return nil
		}
		return CreateSymlink(target.Path, sourcePath)

	case StatusHasFiles:
		if dryRun {
			fmt.Printf("[dry-run] Would migrate files from %s to %s, then create symlink\n", target.Path, sourcePath)
			return nil
		}
		if err := MigrateToSource(target.Path, sourcePath); err != nil {
			return err
		}
		return CreateSymlink(target.Path, sourcePath)

	case StatusConflict:
		link, _ := os.Readlink(target.Path)
		return fmt.Errorf("target is symlink to different location: %s -> %s", target.Path, link)

	case StatusBroken:
		if dryRun {
			fmt.Printf("[dry-run] Would remove broken symlink and recreate: %s\n", target.Path)
			return nil
		}
		os.Remove(target.Path)
		return CreateSymlink(target.Path, sourcePath)

	default:
		return fmt.Errorf("unknown target status: %s", status)
	}
}

// mergeDirectories copies files from src to dst, skipping existing files
func mergeDirectories(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Skip if destination exists
		if _, err := os.Stat(dstPath); err == nil {
			fmt.Printf("  skip (exists): %s\n", relPath)
			return nil
		}

		return copyFile(path, dstPath)
	})
}

// copyDirectory copies a directory recursively
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

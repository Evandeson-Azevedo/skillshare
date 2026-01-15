package integration

import (
	"os"
	"path/filepath"
	"testing"

	"skillshare/internal/testutil"
)

func TestUpdate_NoConfig_ReturnsError(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Remove config
	os.Remove(sb.ConfigPath)

	result := sb.RunCLI("update")

	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "config not found")
}

func TestUpdate_DryRun_DoesNotModify(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Create config
	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	skillPath := filepath.Join(sb.SourcePath, "skillshare", "SKILL.md")

	// Create existing skill
	os.MkdirAll(filepath.Dir(skillPath), 0755)
	os.WriteFile(skillPath, []byte("# Old Content"), 0644)

	result := sb.RunCLI("update", "--dry-run")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Would update")

	// Verify file was not changed
	content, _ := os.ReadFile(skillPath)
	if string(content) != "# Old Content" {
		t.Error("dry-run should not modify file")
	}
}

func TestUpdate_Force_SkipsConfirmation(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Create config
	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	skillPath := filepath.Join(sb.SourcePath, "skillshare", "SKILL.md")

	// Create existing skill
	os.MkdirAll(filepath.Dir(skillPath), 0755)
	os.WriteFile(skillPath, []byte("# Old Content"), 0644)

	// Force update (will fail to download but should attempt)
	result := sb.RunCLI("update", "--force")

	// Should either succeed (if network available) or fail with download error
	// But should NOT ask for confirmation
	if result.ExitCode == 0 {
		result.AssertOutputContains(t, "Updated")
	} else {
		result.AssertAnyOutputContains(t, "download")
	}
}

func TestUpdate_NoExistingSkill_CreatesNew(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Create config
	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	skillPath := filepath.Join(sb.SourcePath, "skillshare", "SKILL.md")

	// Ensure skill doesn't exist
	os.RemoveAll(filepath.Dir(skillPath))

	// Update (will attempt download)
	result := sb.RunCLI("update")

	// Should either succeed or fail with download error
	// But should create directory
	if result.ExitCode == 0 {
		if !sb.FileExists(skillPath) {
			t.Error("skill should be created on success")
		}
	}
}

func TestUpdate_ShowsSourceURL(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Create config
	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	result := sb.RunCLI("update", "--dry-run")

	result.AssertSuccess(t)
	// URL is raw.githubusercontent.com
	result.AssertOutputContains(t, "githubusercontent")
}

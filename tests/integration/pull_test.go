package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"skillshare/internal/testutil"
)

func TestPull_FindsLocalSkills(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	targetPath := sb.CreateTarget("claude")

	// Create local skill in target (not a symlink)
	localSkillPath := filepath.Join(targetPath, "local-skill")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# Local Skill"), 0644)

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets:
  claude:
    path: ` + targetPath + `
`)

	// Run with --dry-run to just see what would be pulled
	result := sb.RunCLI("pull", "--dry-run")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "local-skill")
}

func TestPull_SpecificTarget_OnlyThat(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	claudePath := sb.CreateTarget("claude")
	codexPath := sb.CreateTarget("codex")

	// Create local skills in both targets
	os.MkdirAll(filepath.Join(claudePath, "claude-skill"), 0755)
	os.WriteFile(filepath.Join(claudePath, "claude-skill", "SKILL.md"), []byte("# Claude"), 0644)
	os.MkdirAll(filepath.Join(codexPath, "codex-skill"), 0755)
	os.WriteFile(filepath.Join(codexPath, "codex-skill", "SKILL.md"), []byte("# Codex"), 0644)

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets:
  claude:
    path: ` + claudePath + `
  codex:
    path: ` + codexPath + `
`)

	result := sb.RunCLI("pull", "claude", "--dry-run")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "claude-skill")
	result.AssertOutputNotContains(t, "codex-skill")
}

func TestPull_All_FromAllTargets(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	claudePath := sb.CreateTarget("claude")
	codexPath := sb.CreateTarget("codex")

	// Create local skills in both targets
	os.MkdirAll(filepath.Join(claudePath, "claude-skill"), 0755)
	os.WriteFile(filepath.Join(claudePath, "claude-skill", "SKILL.md"), []byte("# Claude"), 0644)
	os.MkdirAll(filepath.Join(codexPath, "codex-skill"), 0755)
	os.WriteFile(filepath.Join(codexPath, "codex-skill", "SKILL.md"), []byte("# Codex"), 0644)

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets:
  claude:
    path: ` + claudePath + `
  codex:
    path: ` + codexPath + `
`)

	result := sb.RunCLI("pull", "--all", "--dry-run")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "claude-skill")
	result.AssertOutputContains(t, "codex-skill")
}

func TestPull_NoLocalSkills_ShowsMessage(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.CreateSkill("shared-skill", map[string]string{"SKILL.md": "# Shared"})
	targetPath := sb.CreateTarget("claude")

	// Create only symlinked skill (not local)
	os.Symlink(filepath.Join(sb.SourcePath, "shared-skill"), filepath.Join(targetPath, "shared-skill"))

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets:
  claude:
    path: ` + targetPath + `
`)

	result := sb.RunCLI("pull")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "No local skills")
}

func TestPull_TargetNotFound_ReturnsError(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	result := sb.RunCLI("pull", "nonexistent")

	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "not found")
}

func TestPull_CopiesToSource(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	targetPath := sb.CreateTarget("claude")

	// Create local skill
	localSkillPath := filepath.Join(targetPath, "new-skill")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# New Skill Content"), 0644)

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets:
  claude:
    path: ` + targetPath + `
`)

	// Run with force to skip confirmation
	result := sb.RunCLIWithInput("y\n", "pull", "--force")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "copied to source")

	// Verify skill was copied to source
	copiedSkillPath := filepath.Join(sb.SourcePath, "new-skill", "SKILL.md")
	if !sb.FileExists(copiedSkillPath) {
		t.Error("skill should be copied to source")
	}
}

func TestPull_ExistsInSource_Skips(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Create skill in source first
	sb.CreateSkill("existing-skill", map[string]string{"SKILL.md": "# Source Version"})

	targetPath := sb.CreateTarget("claude")

	// Create same skill in target (local copy)
	localSkillPath := filepath.Join(targetPath, "existing-skill")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# Target Version"), 0644)

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets:
  claude:
    path: ` + targetPath + `
`)

	result := sb.RunCLIWithInput("y\n", "pull")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "skipped")
}

func TestPull_MultipleTargets_RequiresAllOrName(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	claudePath := sb.CreateTarget("claude")
	codexPath := sb.CreateTarget("codex")

	// Create local skills in both targets
	os.MkdirAll(filepath.Join(claudePath, "skill1"), 0755)
	os.WriteFile(filepath.Join(claudePath, "skill1", "SKILL.md"), []byte("# 1"), 0644)
	os.MkdirAll(filepath.Join(codexPath, "skill2"), 0755)
	os.WriteFile(filepath.Join(codexPath, "skill2", "SKILL.md"), []byte("# 2"), 0644)

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets:
  claude:
    path: ` + claudePath + `
  codex:
    path: ` + codexPath + `
`)

	// Without specifying target or --all
	result := sb.RunCLI("pull")

	// Should ask to specify target
	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Specify a target")
}

// Tests for --remote flag

func TestPull_Remote_NoGitRepo_ShowsError(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	result := sb.RunCLI("pull", "--remote")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "not a git repository")
}

func TestPull_Remote_NoRemote_ShowsError(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Initialize git but no remote
	cmd := exec.Command("git", "init")
	cmd.Dir = sb.SourcePath
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	result := sb.RunCLI("pull", "--remote")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "No git remote")
}

func TestPull_Remote_UncommittedChanges_Refuses(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create bare repo as "remote"
	bareRepo := filepath.Join(sb.Home, "remote.git")
	cmd := exec.Command("git", "init", "--bare", bareRepo)
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	// Initialize git and add remote
	cmd = exec.Command("git", "init")
	cmd.Dir = sb.SourcePath
	cmd.Run()

	cmd = exec.Command("git", "remote", "add", "origin", bareRepo)
	cmd.Dir = sb.SourcePath
	cmd.Run()

	configGitForPull(t, sb.SourcePath)

	// Initial commit and push
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "initial")
	cmd.Dir = sb.SourcePath
	cmd.Run()

	cmd = exec.Command("git", "push", "-u", "origin", "master")
	cmd.Dir = sb.SourcePath
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = sb.SourcePath
		cmd.Run()
	}

	// Create uncommitted changes
	sb.CreateSkill("uncommitted-skill", map[string]string{"SKILL.md": "# Uncommitted"})

	result := sb.RunCLI("pull", "--remote")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Local changes detected")
}

func TestPull_Remote_DryRun_ShowsActions(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create bare repo as "remote"
	bareRepo := filepath.Join(sb.Home, "remote.git")
	cmd := exec.Command("git", "init", "--bare", bareRepo)
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	// Initialize git and add remote
	cmd = exec.Command("git", "init")
	cmd.Dir = sb.SourcePath
	cmd.Run()

	cmd = exec.Command("git", "remote", "add", "origin", bareRepo)
	cmd.Dir = sb.SourcePath
	cmd.Run()

	configGitForPull(t, sb.SourcePath)

	// Initial commit and push
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "initial")
	cmd.Dir = sb.SourcePath
	cmd.Run()

	cmd = exec.Command("git", "push", "-u", "origin", "master")
	cmd.Dir = sb.SourcePath
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = sb.SourcePath
		cmd.Run()
	}

	result := sb.RunCLI("pull", "--remote", "--dry-run")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "dry-run")
}

func TestPull_Remote_ActualPull_AndSyncs(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	targetPath := sb.CreateTarget("claude")

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets:
  claude:
    path: ` + targetPath + `
`)

	// Create bare repo as "remote"
	bareRepo := filepath.Join(sb.Home, "remote.git")
	cmd := exec.Command("git", "init", "--bare", bareRepo)
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	// Initialize git and add remote
	cmd = exec.Command("git", "init")
	cmd.Dir = sb.SourcePath
	cmd.Run()

	cmd = exec.Command("git", "remote", "add", "origin", bareRepo)
	cmd.Dir = sb.SourcePath
	cmd.Run()

	configGitForPull(t, sb.SourcePath)

	// Create skill, commit, and push
	sb.CreateSkill("remote-skill", map[string]string{"SKILL.md": "# Remote Skill"})

	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = sb.SourcePath
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "add skill")
	cmd.Dir = sb.SourcePath
	cmd.Run()

	cmd = exec.Command("git", "push", "-u", "origin", "master")
	cmd.Dir = sb.SourcePath
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = sb.SourcePath
		cmd.Run()
	}

	// Now run pull --remote (already up to date, but should sync)
	result := sb.RunCLI("pull", "--remote")

	result.AssertSuccess(t)
	// Should sync to target
	if !sb.FileExists(filepath.Join(targetPath, "remote-skill", "SKILL.md")) {
		t.Error("skill should be synced to target after pull --remote")
	}
}

// Helper function for pull tests
func configGitForPull(t *testing.T, dir string) {
	cmd := exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = dir
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = dir
	cmd.Run()
}

package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeployDotfiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dotfiles-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")

	// Create dummy source dotfile
	if err := os.MkdirAll(filepath.Join(srcDir, "hypr"), 0755); err != nil {
		t.Fatalf("Failed to create source hypr dir: %v", err)
	}
	sampleContent := "hyprland config sample"
	if err := os.WriteFile(filepath.Join(srcDir, "hypr", "hyprland.conf"), []byte(sampleContent), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create initial target file to test backup logic
	if err := os.MkdirAll(filepath.Join(targetDir, "hypr"), 0755); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "hypr", "old.conf"), []byte("old config"), 0644); err != nil {
		t.Fatalf("Failed to write old target file: %v", err)
	}

	var logs []string
	logCb := func(msg string) {
		logs = append(logs, msg)
	}

	// Test Real Deploy (DryRun = false)
	err = DeployDotfiles(srcDir, targetDir, false, logCb)
	if err != nil {
		t.Fatalf("DeployDotfiles returned error: %v", err)
	}

	// Verify file was copied to target
	copiedFile := filepath.Join(targetDir, "hypr", "hyprland.conf")
	data, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Errorf("Expected copied file at '%s', got error: %v", copiedFile, err)
	}
	if string(data) != sampleContent {
		t.Errorf("Expected content '%s', got '%s'", sampleContent, string(data))
	}

	if len(logs) == 0 {
		t.Errorf("Expected logs during deployment, got 0")
	}
}

func TestDeployDotfilesDryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dotfiles-dryrun-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")

	if err := os.MkdirAll(filepath.Join(srcDir, "kitty"), 0755); err != nil {
		t.Fatalf("Failed to create source kitty dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "kitty", "kitty.conf"), []byte("kitty test"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	var logs []string
	logCb := func(msg string) {
		logs = append(logs, msg)
	}

	// Test DryRun Deploy (DryRun = true)
	err = DeployDotfiles(srcDir, targetDir, true, logCb)
	if err != nil {
		t.Fatalf("DeployDotfiles DryRun returned error: %v", err)
	}

	// Target file should NOT exist in DryRun mode
	copiedFile := filepath.Join(targetDir, "kitty", "kitty.conf")
	if _, err := os.Stat(copiedFile); !os.IsNotExist(err) {
		t.Errorf("Expected file NOT to be created in DryRun mode, but it exists at '%s'", copiedFile)
	}
}

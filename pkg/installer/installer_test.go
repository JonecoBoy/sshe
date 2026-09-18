package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunnerDryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "arch-setup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	targetConfig := filepath.Join(tmpDir, ".config")
	dotfilesSrc := filepath.Join(tmpDir, "dotfiles", ".config", "testpkg")

	if err := os.MkdirAll(dotfilesSrc, 0755); err != nil {
		t.Fatalf("Failed to create test dotfiles directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dotfilesSrc, "config.conf"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var logged []string
	logCb := func(msg string) {
		logged = append(logged, msg)
	}

	runner := NewRunner(tmpDir, targetConfig, true, logCb)
	runner.PacmanPackages = []string{"firefox", "kitty"}
	runner.YayPackages = []string{"brave-bin"}

	steps := runner.GetDefaultSteps()
	for _, step := range steps {
		err := runner.RunStep(step.ID)
		if err != nil {
			t.Errorf("Step '%s' failed in DryRun mode: %v", step.ID, err)
		}
	}

	if len(logged) == 0 {
		t.Errorf("Expected logs during DryRun execution, got 0")
	}
}

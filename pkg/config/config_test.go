package config

import (
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configPath := filepath.Join("..", "..", "config", "packages.yaml")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load packages.yaml: %v", err)
	}

	if len(cfg.MustHave.Pacman) == 0 {
		t.Errorf("Expected pacman packages in MustHave, got 0")
	}

	if len(cfg.Categories) == 0 {
		t.Errorf("Expected configured categories, got 0")
	}

	pacmanPkgs, yayPkgs, npmPkgs, customCmds := cfg.GetSelectedPackages()
	if len(pacmanPkgs) == 0 {
		t.Errorf("Expected selected pacman packages, got 0")
	}
	if len(yayPkgs) == 0 {
		t.Errorf("Expected selected yay packages, got 0")
	}
	if len(customCmds) == 0 {
		t.Errorf("Expected selected custom commands (Codex/Antigravity), got 0")
	}
	t.Logf("NPM packages selected by default: %d", len(npmPkgs))

	// Verify 'firefox' (default single option in browsers category) is present in pacmanPkgs
	foundFirefox := false
	for _, p := range pacmanPkgs {
		if p == "firefox" {
			foundFirefox = true
			break
		}
	}
	if !foundFirefox {
		t.Errorf("Expected 'firefox' selected by default in browsers category")
	}
}

package tui

import (
	"path/filepath"
	"testing"

	"sshe/pkg/config"
	"sshe/pkg/installer"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIStateTransitions(t *testing.T) {
	configPath := filepath.Join("..", "..", "config", "packages.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config for TUI test: %v", err)
	}

	runner := installer.NewRunner(".", "/tmp/.config_test", true, nil)
	model := NewModel(cfg, nil, runner, true)

	// 1. Initial State should be StateWelcome
	if model.State != StateWelcome {
		t.Errorf("Expected initial state StateWelcome (%d), got %d", StateWelcome, model.State)
	}

	// 2. Pressing 'Enter' on Welcome screen transitions to StateWizard
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m := updatedModel.(Model)
	if m.State != StateWizard {
		t.Errorf("Expected transition to StateWizard (%d), got %d", StateWizard, m.State)
	}

	// 3. Pressing 'j' or 'down' increases OptionIndex
	initialOptIndex := m.OptionIndex
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(Model)
	if m.OptionIndex != initialOptIndex+1 {
		t.Errorf("Expected OptionIndex %d after Down key, got %d", initialOptIndex+1, m.OptionIndex)
	}

	// 4. Pressing 'b' returns to StateWelcome
	m.CategoryIndex = 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m = updatedModel.(Model)
	if m.State != StateWelcome {
		t.Errorf("Expected transition back to StateWelcome (%d), got %d", StateWelcome, m.State)
	}

	// 5. Test View() rendering does not panic or crash
	viewOutput := m.View()
	if len(viewOutput) == 0 {
		t.Errorf("Expected non-empty view string output")
	}
}

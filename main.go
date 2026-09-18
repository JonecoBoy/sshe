package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"sshe/pkg/config"
	"sshe/pkg/installer"
	"sshe/pkg/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	dryRunFlag := flag.Bool("dry-run", false, "Run in simulation mode without modifying system files or packages")
	configPathFlag := flag.String("config", filepath.Join("config", "packages.yaml"), "Path to the package configuration YAML file")
	appConfigPathFlag := flag.String("app-config", filepath.Join("config", "config.yaml"), "Path to the app settings & theme configuration YAML file")
	targetConfigFlag := flag.String("target-config", filepath.Join(os.Getenv("HOME"), ".config"), "Path to the target configuration directory (~/.config)")
	flag.Parse()

	// 1. Load package configuration
	cfg, err := config.LoadConfig(*configPathFlag)
	if err != nil {
		fmt.Printf("Error loading package configuration file '%s': %v\n", *configPathFlag, err)
		os.Exit(1)
	}

	// 2. Load app configuration & theme
	appCfg, _ := config.LoadAppConfig(*appConfigPathFlag)

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	var p *tea.Program

	// Callback to pipe runner logs directly into the TUI viewport
	logFunc := func(logLine string) {
		if p != nil {
			p.Send(tui.LogMsg(logLine))
		}
	}

	runner := installer.NewRunner(workDir, *targetConfigFlag, *dryRunFlag, logFunc)
	model := tui.NewModel(cfg, appCfg, runner, *dryRunFlag)

	p = tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI application: %v\n", err)
		os.Exit(1)
	}
}

// LogMsg helper for type casting
func LogMsg(s string) tea.Msg {
	return tui.LogMsg(s)
}

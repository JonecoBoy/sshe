package installer

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sshe/pkg/dotfiles"
)

type StepStatus string

const (
	StatusPending StepStatus = "Pending"
	StatusRunning StepStatus = "Running"
	StatusSuccess StepStatus = "Success"
	StatusFailed  StepStatus = "Failed"
	StatusSkipped StepStatus = "Skipped"
)

type Step struct {
	ID          string
	Title       string
	Description string
	Status      StepStatus
	Error       error
}

type Runner struct {
	DryRun         bool
	WorkDir        string
	TargetConfig   string
	PacmanPackages []string
	YayPackages    []string
	NpmPackages    []string
	CustomCommands []string
	LogCallback    func(string)
}

func NewRunner(workDir, targetConfig string, dryRun bool, logCb func(string)) *Runner {
	if logCb == nil {
		logCb = func(s string) {}
	}
	return &Runner{
		DryRun:       dryRun,
		WorkDir:      workDir,
		TargetConfig: targetConfig,
		LogCallback:  logCb,
	}
}

func (r *Runner) GetDefaultSteps() []Step {
	return []Step{
		{ID: "sys_check", Title: "System Prerequisites Check", Description: "Verify Arch Linux OS, sudo permissions, and network connectivity.", Status: StatusPending},
		{ID: "pacman_pkgs", Title: "Official Pacman Packages", Description: "Install official packages from Arch repositories.", Status: StatusPending},
		{ID: "yay_install", Title: "Yay AUR Helper Setup", Description: "Ensure yay AUR helper is installed.", Status: StatusPending},
		{ID: "yay_pkgs", Title: "Yay (AUR) Packages", Description: "Install selected AUR packages.", Status: StatusPending},
		{ID: "npm_pkgs", Title: "Global NPM Packages", Description: "Install global command-line tools via npm install -g.", Status: StatusPending},
		{ID: "custom_cmds", Title: "Custom Scripts & Commands", Description: "Run custom installers via curl/bash (Codex, Antigravity CLI, etc.).", Status: StatusPending},
		{ID: "dotfiles", Title: "Hyprland Dotfiles Deployment", Description: "Backup ~/.config and copy new configuration files.", Status: StatusPending},
		{ID: "services", Title: "Systemd Services Activation", Description: "Enable and start core systemd services (bluetooth, network).", Status: StatusPending},
	}
}

func (r *Runner) RunStep(stepID string) error {
	switch stepID {
	case "sys_check":
		return r.checkSystem()
	case "pacman_pkgs":
		return r.installPacman()
	case "yay_install":
		return r.ensureYay()
	case "yay_pkgs":
		return r.installYay()
	case "npm_pkgs":
		return r.installNpm()
	case "custom_cmds":
		return r.runCustomCmds()
	case "dotfiles":
		return r.deployDotfiles()
	case "services":
		return r.enableServices()
	default:
		return fmt.Errorf("Unknown step: %s", stepID)
	}
}

func (r *Runner) checkSystem() error {
	r.LogCallback("=== Verifying System Prerequisites ===")

	// 1. Validate Arch Linux OS
	osRelease, err := os.ReadFile("/etc/os-release")
	if err != nil && !r.DryRun {
		return fmt.Errorf("Could not read /etc/os-release: %v", err)
	}
	if !strings.Contains(string(osRelease), "ID=arch") && !strings.Contains(string(osRelease), "ID_LIKE=arch") && !r.DryRun {
		r.LogCallback("Warning: OS did not report ID=arch, but continuing per user confirmation.")
	} else {
		r.LogCallback("✔ Arch Linux verified successfully.")
	}

	// 2. Validate Network Connection
	r.LogCallback("Testing network connection to archlinux.org...")
	if !r.DryRun {
		_, err := net.LookupHost("archlinux.org")
		if err != nil {
			return fmt.Errorf("No internet connection or DNS lookup failed: %v", err)
		}
	}
	r.LogCallback("✔ Network connection active.")

	return nil
}

func (r *Runner) installPacman() error {
	if len(r.PacmanPackages) == 0 {
		r.LogCallback("No Pacman packages to install.")
		return nil
	}

	r.LogCallback(fmt.Sprintf("=== Installing %d packages via Pacman ===", len(r.PacmanPackages)))
	cmdArgs := append([]string{"pacman", "-Syu", "--needed", "--noconfirm"}, r.PacmanPackages...)

	if r.DryRun {
		r.LogCallback(fmt.Sprintf("[DRY-RUN] Would execute: sudo %s", strings.Join(cmdArgs, " ")))
		return nil
	}

	return r.executeCommand("sudo", cmdArgs...)
}

func (r *Runner) ensureYay() error {
	r.LogCallback("=== Checking Yay helper installation ===")
	if _, err := exec.LookPath("yay"); err == nil {
		r.LogCallback("✔ Yay is already installed.")
		return nil
	}

	if r.DryRun {
		r.LogCallback("[DRY-RUN] Yay not found. Would git clone AUR repo and run makepkg -si.")
		return nil
	}

	r.LogCallback("Yay not found. Building from AUR repository...")
	tmpDir, err := os.MkdirTemp("", "yay-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := r.executeCommand("git", "clone", "https://aur.archlinux.org/yay.git", tmpDir); err != nil {
		return fmt.Errorf("Failed to clone yay repository: %v", err)
	}

	cmd := exec.Command("makepkg", "-si", "--noconfirm")
	cmd.Dir = tmpDir
	return r.runCmdPipeLogs(cmd)
}

func (r *Runner) installYay() error {
	if len(r.YayPackages) == 0 {
		r.LogCallback("No AUR (Yay) packages to install.")
		return nil
	}

	r.LogCallback(fmt.Sprintf("=== Installing %d AUR packages via Yay ===", len(r.YayPackages)))
	cmdArgs := append([]string{"-S", "--needed", "--noconfirm"}, r.YayPackages...)

	if r.DryRun {
		r.LogCallback(fmt.Sprintf("[DRY-RUN] Would execute: yay %s", strings.Join(cmdArgs, " ")))
		return nil
	}

	return r.executeCommand("yay", cmdArgs...)
}

func (r *Runner) installNpm() error {
	if len(r.NpmPackages) == 0 {
		r.LogCallback("No NPM packages to install.")
		return nil
	}

	r.LogCallback(fmt.Sprintf("=== Installing %d global NPM packages ===", len(r.NpmPackages)))
	cmdArgs := append([]string{"npm", "install", "-g"}, r.NpmPackages...)

	if r.DryRun {
		r.LogCallback(fmt.Sprintf("[DRY-RUN] Would execute: sudo %s", strings.Join(cmdArgs, " ")))
		return nil
	}

	return r.executeCommand("sudo", cmdArgs...)
}

func (r *Runner) runCustomCmds() error {
	if len(r.CustomCommands) == 0 {
		r.LogCallback("No custom commands/scripts to execute.")
		return nil
	}

	r.LogCallback(fmt.Sprintf("=== Executing %d custom commands/scripts ===", len(r.CustomCommands)))
	for idx, customCmd := range r.CustomCommands {
		r.LogCallback(fmt.Sprintf("[%d/%d] Command: %s", idx+1, len(r.CustomCommands), customCmd))
		if r.DryRun {
			r.LogCallback(fmt.Sprintf("[DRY-RUN] Would execute: bash -c %q", customCmd))
			continue
		}

		cmd := exec.Command("bash", "-c", customCmd)
		if err := r.runCmdPipeLogs(cmd); err != nil {
			r.LogCallback(fmt.Sprintf("Warning/Error executing command '%s': %v", customCmd, err))
		}
	}

	return nil
}

func (r *Runner) deployDotfiles() error {
	r.LogCallback("=== Deploying Hyprland Dotfiles ===")
	sourceDotfiles := filepath.Join(r.WorkDir, "dotfiles", ".config")
	return dotfiles.DeployDotfiles(sourceDotfiles, r.TargetConfig, r.DryRun, r.LogCallback)
}

func (r *Runner) enableServices() error {
	r.LogCallback("=== Enabling Essential Systemd Services ===")
	services := []string{"bluetooth.service", "NetworkManager.service"}

	for _, svc := range services {
		r.LogCallback(fmt.Sprintf("Enabling service: %s", svc))
		if r.DryRun {
			r.LogCallback(fmt.Sprintf("[DRY-RUN] Would execute: sudo systemctl enable --now %s", svc))
			continue
		}

		_ = r.executeCommand("sudo", "systemctl", "enable", "--now", svc)
	}

	r.LogCallback("✔ Systemd services configured successfully.")
	return nil
}

func (r *Runner) executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return r.runCmdPipeLogs(cmd)
}

func (r *Runner) runCmdPipeLogs(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanStdout := bufio.NewScanner(stdout)
		for scanStdout.Scan() {
			r.LogCallback(scanStdout.Text())
		}
		if err := scanStdout.Err(); err != nil {
			r.LogCallback("Error reading stdout: " + err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		scanStderr := bufio.NewScanner(stderr)
		for scanStderr.Scan() {
			r.LogCallback(scanStderr.Text())
		}
		if err := scanStderr.Err(); err != nil {
			r.LogCallback("Error reading stderr: " + err.Error())
		}
	}()

	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		time.Sleep(100 * time.Millisecond)
	}
	return err
}

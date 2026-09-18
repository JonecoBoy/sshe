# SSHE - Super Simple Hyper Env

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![UI Framework](https://img.shields.io/badge/TUI-Bubble%20Tea-FAFAFA?style=flat&logo=charm)](https://github.com/charmbracelet/bubbletea)
[![OS](https://img.shields.io/badge/OS-Arch%20Linux-1793D1?style=flat&logo=archlinux)](https://archlinux.org)

**SSHE** (Super Simple Hyper Env) is a modern, terminal-based setup tool for Arch Linux built with **Go** and **Bubble Tea**. It automates the installation and configuration of a complete Hyprland desktop environment, dotfiles, audio servers, fonts, and developer tools through an interactive terminal interface (TUI).

---

## Key Features

- **Interactive TUI Wizard**: Built using [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) for a modern terminal interface.
- **Declarative YAML Configurations**: Easy modification of packages (`config/packages.yaml`) and TUI appearance (`config/config.yaml`).
- **Hyprland Desktop Environment**: Automatically installs core desktop components including Hyprland, Hyprpaper, Hyprlock, Hypridle, Waybar, Rofi, Pipewire, and essential fonts/icons.
- **Customizable Software Selection**: Choose your preferred browsers, terminal emulators, shells, file managers, code editors, container runtimes, and themes (single-choice and multi-choice options).
- **Simulation Mode (`--dry-run`)**: Test and preview installation steps without modifying system packages or files.
- **Dotfile Management**: Automatically links and deploys user configuration dotfiles into `~/.config`.

---

## Repository Structure

```
.
├── main.go               # Application entry point
├── config/
│   ├── config.yaml       # App metadata and TUI color theme settings
│   └── packages.yaml     # Mandatory packages and category wizard definitions
├── pkg/
│   ├── config/           # YAML config parser and data structs
│   ├── dotfiles/         # Dotfile linking logic
│   ├── installer/        # Package manager & step execution runner
│   └── tui/              # Bubble Tea TUI components, layout & styling
├── dotfiles/             # Default configuration templates (.config)
├── go.mod
└── go.sum
```

---

## Getting Started

### Option 1: Download Pre-built Binary (Recommended)

You can download the latest pre-built binary directly from the [GitHub Releases](https://github.com/your-username/sshe/releases) page:

```bash
# Download the latest binary via curl
curl -sL https://github.com/your-username/sshe/releases/latest/download/sshe -o sshe

# Make it executable
chmod +x sshe

# Run in dry-run simulation mode
./sshe --dry-run
```

### Option 2: Build from Source

#### Prerequisites

- **Operating System**: Arch Linux (or Arch-based distribution)
- **Go**: `1.22` or newer
- **Package Managers**: `pacman` and `yay` (for AUR packages)

#### Compilation & Execution

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your-username/sshe.git
   cd sshe
   ```

2. **Build the binary**:
   ```bash
   go build -o sshe main.go
   ```

3. **Run in simulation mode (Dry Run)**:
   ```bash
   ./sshe --dry-run
   ```

4. **Run actual installation**:
   ```bash
   ./sshe
   ```

---

## CLI Options & Flags

| Flag | Default Value | Description |
| :--- | :--- | :--- |
| `--dry-run` | `false` | Run in simulation mode without modifying system files or installing packages |
| `--config` | `config/packages.yaml` | Path to package selection schema YAML file |
| `--app-config` | `config/config.yaml` | Path to app theme and UI settings YAML file |
| `--target-config` | `$HOME/.config` | Target directory for deploying dotfiles |

Example with custom paths:
```bash
./sshe --config custom_packages.yaml --target-config ~/.config --dry-run
```

---

## Configuration

### 1. `config/packages.yaml`
Defines packages installed by default (`must_have`) and interactive selection categories (`categories`).

```yaml
must_have:
  pacman:
    - base-devel
    - git
    - hyprland
    - pipewire
  yay:
    - swww

categories:
  - id: "browsers"
    title: "Web Browsers"
    description: "Choose your primary web browser."
    type: "single"
    options:
      - id: "firefox"
        name: "Firefox (Recommended)"
        default: true
        pacman: ["firefox"]
```

### 2. `config/config.yaml`
Defines application metadata, welcome screen texts, and TUI color customization (supporting hex color codes):

```yaml
# SSHE Application Configuration Settings
app:
  name: "SSHE"
  subtitle: "Super Simple Hyper Env"
  operating_system: "Arch Linux"
  welcome_title: "Welcome to SSHE (Super Simple Hyper Env) Setup Tool!"
  welcome_subtitle: "This tool configures your Arch Linux system with Hyprland, Waybar, Rofi, Pipewire Audio, themes, and your favorite apps."

theme:
  banner_fg: "#CD201F"
  title_fg: "#89B4FA"
  title_bg: "#1E1E2E"
  subtitle_fg: "#BAC2DE"
  selected_fg: "#A6E3A1"
  unselected_fg: "#CDD6F4"
  description_fg: "#6C7086"
  border_fg: "#89B4FA"
  status_pending: "#6C7086"
  status_running: "#F9E2AF"
  status_success: "#A6E3A1"
  status_error: "#F38BA8"
```

---

## 📂 Dotfiles Deployment (`dotfiles/.config`)

SSHE automatically manages and deploys your user configuration files:

- **Source Directory**: `dotfiles/.config/` in the repository root.
- **Target Directory**: `~/.config/` (configurable via `--target-config`).

### How Dotfiles Deployment Works:
1. **Automatic Backup**: Before copying, SSHE creates a timestamped backup copy of your existing `~/.config` directory (e.g., `~/.config_backup_20260918_012500`).
2. **Recursive Copying**: Every folder and file placed inside `dotfiles/.config/` are recursively copied into `~/.config/`.
3. **Dry-Run Preview**: When running with `--dry-run`, SSHE logs every file that would be backed up and copied without modifying any system or configuration files.

---

## 🗺️ Roadmap

- [x] **Arch Linux** support (`pacman` & `yay` / AUR)
- [ ] Multi-OS & Package Manager expansion:
  - [ ] **Ubuntu / Debian**: `apt`
  - [ ] **openSUSE**: `zypper`
  - [ ] **FreeBSD**: `pkg`
  - [ ] **Fedora / RHEL**: `dnf`
  - [ ] **Void Linux**: `xbps`
- [ ] *Maybe / Under Consideration*:
  - [ ] **macOS**: `brew` (Homebrew for cross-platform dotfile management)
  - [ ] **Alpine Linux**: `apk`
  - [ ] **Nix / NixOS**: `nix`
  - [ ] Custom post-install hook scripts in YAML
  - [ ] Profile export / import system

---

## License

Distributed under the GNU License. See `LICENSE` for more information.

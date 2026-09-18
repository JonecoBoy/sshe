package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppSettings struct {
	Name            string `yaml:"name"`
	Subtitle        string `yaml:"subtitle"`
	OperatingSystem string `yaml:"operating_system"`
	WelcomeTitle    string `yaml:"welcome_title"`
	WelcomeSubtitle string `yaml:"welcome_subtitle"`
}

type ThemeSettings struct {
	BannerFg      string `yaml:"banner_fg"`
	TitleFg       string `yaml:"title_fg"`
	TitleBg       string `yaml:"title_bg"`
	SubtitleFg    string `yaml:"subtitle_fg"`
	SelectedFg    string `yaml:"selected_fg"`
	UnselectedFg  string `yaml:"unselected_fg"`
	DescriptionFg string `yaml:"description_fg"`
	BorderFg      string `yaml:"border_fg"`
	StatusPending string `yaml:"status_pending"`
	StatusRunning string `yaml:"status_running"`
	StatusSuccess string `yaml:"status_success"`
	StatusError   string `yaml:"status_error"`
}

type AppConfig struct {
	App   AppSettings   `yaml:"app"`
	Theme ThemeSettings `yaml:"theme"`
}

func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		App: AppSettings{
			Name:            "SSHE",
			Subtitle:        "Super Simple Hyper Env",
			OperatingSystem: "Arch Linux",
			WelcomeTitle:    "Welcome to SSHE (Super Simple Hyper Env) Setup Tool!",
			WelcomeSubtitle: "This tool configures your Arch Linux system with Hyprland, Waybar, Rofi, Pipewire Audio, themes, and your favorite apps.",
		},
		Theme: ThemeSettings{
			BannerFg:      "#CD201F",
			TitleFg:       "#89B4FA",
			TitleBg:       "#1E1E2E",
			SubtitleFg:    "#BAC2DE",
			SelectedFg:    "#A6E3A1",
			UnselectedFg:  "#CDD6F4",
			DescriptionFg: "#6C7086",
			BorderFg:      "#89B4FA",
			StatusPending: "#6C7086",
			StatusRunning: "#F9E2AF",
			StatusSuccess: "#A6E3A1",
			StatusError:   "#F38BA8",
		},
	}
}

func LoadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultAppConfig(), nil
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultAppConfig(), nil
	}
	return &cfg, nil
}

type MustHave struct {
	Pacman  []string `yaml:"pacman"`
	Yay     []string `yaml:"yay"`
	Npm     []string `yaml:"npm"`
	Command []string `yaml:"command"`
}

type Option struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Default     bool     `yaml:"default"`
	Pacman      []string `yaml:"pacman"`
	Yay         []string `yaml:"yay"`
	Npm         []string `yaml:"npm"`
	Command     []string `yaml:"command"`
	Selected    bool     `yaml:"-"` // Runtime selection state
}

type Category struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"` // "single" or "multi"
	Options     []Option `yaml:"options"`
}

type Config struct {
	MustHave   MustHave   `yaml:"must_have"`
	Categories []Category `yaml:"categories"`
}

// LoadConfig reads and parses the package configuration YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Initialize runtime Selected state based on Default field
	for i := range cfg.Categories {
		cat := &cfg.Categories[i]
		singleSelected := false
		for j := range cat.Options {
			opt := &cat.Options[j]
			if cat.Type == "single" {
				if opt.Default && !singleSelected {
					opt.Selected = true
					singleSelected = true
				} else {
					opt.Selected = false
				}
			} else {
				opt.Selected = opt.Default
			}
		}
		// If no option has default set in a single-choice category, select the first one by default
		if cat.Type == "single" && !singleSelected && len(cat.Options) > 0 {
			cat.Options[0].Selected = true
		}
	}

	return &cfg, nil
}

// GetSelectedPackages aggregates all selected pacman, yay, npm packages and custom commands without duplicates.
func (c *Config) GetSelectedPackages() (pacmanPkgs, yayPkgs, npmPkgs, customCmds []string) {
	pacmanSet := make(map[string]bool)
	yaySet := make(map[string]bool)
	npmSet := make(map[string]bool)

	// Add MustHave packages
	for _, p := range c.MustHave.Pacman {
		if !pacmanSet[p] {
			pacmanSet[p] = true
			pacmanPkgs = append(pacmanPkgs, p)
		}
	}
	for _, p := range c.MustHave.Yay {
		if !yaySet[p] {
			yaySet[p] = true
			yayPkgs = append(yayPkgs, p)
		}
	}
	for _, p := range c.MustHave.Npm {
		if !npmSet[p] {
			npmSet[p] = true
			npmPkgs = append(npmPkgs, p)
		}
	}
	customCmds = append(customCmds, c.MustHave.Command...)

	// Add Selected Category Options
	for _, cat := range c.Categories {
		for _, opt := range cat.Options {
			if opt.Selected {
				for _, p := range opt.Pacman {
					if !pacmanSet[p] {
						pacmanSet[p] = true
						pacmanPkgs = append(pacmanPkgs, p)
					}
				}
				for _, p := range opt.Yay {
					if !yaySet[p] {
						yaySet[p] = true
						yayPkgs = append(yayPkgs, p)
					}
				}
				for _, p := range opt.Npm {
					if !npmSet[p] {
						npmSet[p] = true
						npmPkgs = append(npmPkgs, p)
					}
				}
				customCmds = append(customCmds, opt.Command...)
			}
		}
	}

	return pacmanPkgs, yayPkgs, npmPkgs, customCmds
}

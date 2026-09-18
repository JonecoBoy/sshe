package tui

import (
	"fmt"
	"strings"
	"time"

	"sshe/pkg/config"
	"sshe/pkg/installer"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewState int

const (
	StateWelcome ViewState = iota
	StateWizard
	StateConfirm
	StatePipeline
	StateDone
)

type LogMsg string
type stepDoneMsg struct {
	stepIndex int
	err       error
}

type Model struct {
	State         ViewState
	Config        *config.Config
	AppConfig     *config.AppConfig
	Runner        *installer.Runner
	CategoryIndex int
	OptionIndex   int
	Steps         []installer.Step
	CurrentStep   int
	Logs          []string
	Viewport      viewport.Model
	Spinner       spinner.Model
	Err           error
	DryRun        bool
	WindowWidth   int
	WindowHeight  int

	// Dynamic Styles
	TitleStyle      lipgloss.Style
	SubtitleStyle   lipgloss.Style
	SelectedStyle   lipgloss.Style
	UnselectedStyle lipgloss.Style
	DescStyle       lipgloss.Style
	BoxStyle        lipgloss.Style
	StatusPending   string
	StatusRunning   string
	StatusSuccess   string
	StatusError     string
	BannerStyle     lipgloss.Style
}

func NewModel(cfg *config.Config, appCfg *config.AppConfig, runner *installer.Runner, dryRun bool) Model {
	if appCfg == nil {
		appCfg = config.DefaultAppConfig()
	}

	theme := appCfg.Theme

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.TitleFg)).
		Background(lipgloss.Color(theme.TitleBg)).
		Padding(0, 1).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SubtitleFg)).
		Italic(true)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.SelectedFg)).
		PaddingLeft(2)

	unselectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.UnselectedFg)).
		PaddingLeft(2)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.DescriptionFg)).
		PaddingLeft(6)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.BorderFg)).
		Padding(1, 2)

	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.BannerFg))

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TitleFg))

	vp := viewport.New(80, 10)
	vp.SetContent("Waiting to start setup process...\n")

	return Model{
		State:           StateWelcome,
		Config:          cfg,
		AppConfig:       appCfg,
		Runner:          runner,
		Steps:           runner.GetDefaultSteps(),
		Spinner:         s,
		Viewport:        vp,
		DryRun:          dryRun,
		WindowWidth:     80,
		WindowHeight:    24,
		TitleStyle:      titleStyle,
		SubtitleStyle:   subtitleStyle,
		SelectedStyle:   selectedStyle,
		UnselectedStyle: unselectedStyle,
		DescStyle:       descStyle,
		BoxStyle:        boxStyle,
		BannerStyle:     bannerStyle,
		StatusPending:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.StatusPending)).Render("[Pending]"),
		StatusRunning:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.StatusRunning)).Bold(true).Render("[⏳ Running]"),
		StatusSuccess:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.StatusSuccess)).Bold(true).Render("[✔ Completed]"),
		StatusError:     lipgloss.NewStyle().Foreground(lipgloss.Color(theme.StatusError)).Bold(true).Render("[✖ Error]"),
	}
}

func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowWidth = msg.Width
		m.WindowHeight = msg.Height
		m.Viewport.Width = msg.Width - 6
		m.Viewport.Height = 8

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		cmds = append(cmds, cmd)

	case LogMsg:
		m.Logs = append(m.Logs, string(msg))
		if len(m.Logs) > 300 {
			m.Logs = m.Logs[len(m.Logs)-300:]
		}
		m.Viewport.SetContent(strings.Join(m.Logs, "\n"))
		m.Viewport.GotoBottom()

	case stepDoneMsg:
		if msg.err != nil {
			m.Steps[msg.stepIndex].Status = installer.StatusFailed
			m.Steps[msg.stepIndex].Error = msg.err
			m.State = StateDone
			m.Err = msg.err
		} else {
			m.Steps[msg.stepIndex].Status = installer.StatusSuccess
			m.CurrentStep++

			if m.CurrentStep < len(m.Steps) {
				m.Steps[m.CurrentStep].Status = installer.StatusRunning
				cmds = append(cmds, m.runStepCmd(m.CurrentStep))
			} else {
				m.State = StateDone
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.State != StatePipeline {
				return m, tea.Quit
			}
		}

		switch m.State {
		case StateWelcome:
			if msg.String() == "enter" || msg.String() == " " {
				m.State = StateWizard
				m.CategoryIndex = 0
				m.OptionIndex = 0
			}

		case StateWizard:
			cat := &m.Config.Categories[m.CategoryIndex]

			switch msg.String() {
			case "up", "k":
				if m.OptionIndex > 0 {
					m.OptionIndex--
				}
			case "down", "j":
				if m.OptionIndex < len(cat.Options)-1 {
					m.OptionIndex++
				}
			case " ":
				if cat.Type == "single" {
					for i := range cat.Options {
						cat.Options[i].Selected = (i == m.OptionIndex)
					}
				} else {
					cat.Options[m.OptionIndex].Selected = !cat.Options[m.OptionIndex].Selected
				}
			case "enter", "tab":
				if m.CategoryIndex < len(m.Config.Categories)-1 {
					m.CategoryIndex++
					m.OptionIndex = 0
				} else {
					m.State = StateConfirm
				}
			case "b":
				if m.CategoryIndex > 0 {
					m.CategoryIndex--
					m.OptionIndex = 0
				} else {
					m.State = StateWelcome
				}
			}

		case StateConfirm:
			if msg.String() == "enter" {
				m.State = StatePipeline
				m.CurrentStep = 0
				m.Steps[0].Status = installer.StatusRunning

				// Update selected packages on runner
				pacmanPkgs, yayPkgs, npmPkgs, customCmds := m.Config.GetSelectedPackages()
				m.Runner.PacmanPackages = pacmanPkgs
				m.Runner.YayPackages = yayPkgs
				m.Runner.NpmPackages = npmPkgs
				m.Runner.CustomCommands = customCmds

				cmds = append(cmds, m.runStepCmd(0))
			} else if msg.String() == "b" {
				m.State = StateWizard
			}

		case StateDone:
			if msg.String() == "enter" || msg.String() == "q" {
				return m, tea.Quit
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) runStepCmd(stepIdx int) tea.Cmd {
	return func() tea.Msg {
		step := m.Steps[stepIdx]
		err := m.Runner.RunStep(step.ID)
		return stepDoneMsg{stepIndex: stepIdx, err: err}
	}
}

func (m Model) View() string {
	var b strings.Builder

	// Top Banner
	bannerText := fmt.Sprintf("          %s - %s          ", m.AppConfig.App.Name, strings.ToUpper(m.AppConfig.App.Subtitle))
	b.WriteString(m.TitleStyle.Render(m.BannerStyle.Render(bannerText)))
	b.WriteString("\n\n")

	switch m.State {
	case StateWelcome:
		b.WriteString(m.viewWelcome())
	case StateWizard:
		b.WriteString(m.viewWizard())
	case StateConfirm:
		b.WriteString(m.viewConfirm())
	case StatePipeline, StateDone:
		b.WriteString(m.viewPipeline())
	}

	return b.String()
}

func (m Model) viewWelcome() string {
	var b strings.Builder
	b.WriteString(m.TitleStyle.Render(m.AppConfig.App.WelcomeTitle))
	b.WriteString(m.SubtitleStyle.Render(m.AppConfig.App.WelcomeSubtitle))
	b.WriteString("\n\n")

	infoBox := fmt.Sprintf(
		"• Operating System: %s\n"+
			"• Package Config YAML: config/packages.yaml\n"+
			"• App Config YAML: config/config.yaml\n"+
			"• Simulation Mode (--dry-run): %v\n\n"+
			"Press [ENTER] or [SPACE] to start component selection wizard.",
		m.AppConfig.App.OperatingSystem, m.DryRun,
	)

	b.WriteString(m.BoxStyle.Render(infoBox))
	b.WriteString("\n\n")
	b.WriteString(m.SubtitleStyle.Render("[ENTER] Start   [q] Quit"))
	return b.String()
}

func (m Model) viewWizard() string {
	var b strings.Builder
	cat := m.Config.Categories[m.CategoryIndex]

	header := fmt.Sprintf("[%d/%d] %s", m.CategoryIndex+1, len(m.Config.Categories), cat.Title)
	if cat.Type == "single" {
		header += " (Single Choice)"
	} else {
		header += " (Multiple Choice)"
	}

	b.WriteString(m.SelectedStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(m.SubtitleStyle.Render(cat.Description))
	b.WriteString("\n\n")

	for i, opt := range cat.Options {
		cursor := "  "
		if i == m.OptionIndex {
			cursor = "❯ "
		}

		check := "[ ]"
		if cat.Type == "single" {
			check = "( )"
			if opt.Selected {
				check = "(•)"
			}
		} else {
			if opt.Selected {
				check = "[x]"
			}
		}

		line := fmt.Sprintf("%s%s %s", cursor, check, opt.Name)
		if i == m.OptionIndex {
			b.WriteString(m.SelectedStyle.Render(line))
			b.WriteString("\n")
			if opt.Description != "" {
				b.WriteString(m.DescStyle.Render(opt.Description))
				b.WriteString("\n")
			}
		} else {
			b.WriteString(m.UnselectedStyle.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.SubtitleStyle.Render("[↑/↓] Navigate   [SPACE] Toggle   [ENTER/TAB] Next   [b] Back   [q] Quit"))
	return b.String()
}

func (m Model) viewConfirm() string {
	var b strings.Builder
	pacmanPkgs, yayPkgs, npmPkgs, customCmds := m.Config.GetSelectedPackages()

	b.WriteString(m.SelectedStyle.Render("Selected Components Summary"))
	b.WriteString("\n\n")

	summary := fmt.Sprintf(
		"• Official Packages (Pacman): %d packages\n"+
			"• AUR Packages (Yay): %d packages\n"+
			"• Global Packages (NPM): %d packages\n"+
			"• Custom Scripts/Commands: %d commands\n"+
			"• Target Dotfiles Destination: %s\n"+
			"• Automatic Backup: ~/.config_backup_<date>\n"+
			"• Simulation Mode (--dry-run): %v",
		len(pacmanPkgs), len(yayPkgs), len(npmPkgs), len(customCmds), m.Runner.TargetConfig, m.DryRun,
	)

	b.WriteString(m.BoxStyle.Render(summary))
	b.WriteString("\n\n")

	b.WriteString(m.SubtitleStyle.Render("Press [ENTER] to start installation or [b] to adjust choices."))
	return b.String()
}

func (m Model) viewPipeline() string {
	var b strings.Builder

	b.WriteString(m.TitleStyle.Render(fmt.Sprintf("%s Installation Pipeline", m.AppConfig.App.Name)))
	b.WriteString("\n\n")

	for _, step := range m.Steps {
		var stStr string
		switch step.Status {
		case installer.StatusPending:
			stStr = m.StatusPending
		case installer.StatusRunning:
			stStr = fmt.Sprintf("%s %s", m.Spinner.View(), m.StatusRunning)
		case installer.StatusSuccess:
			stStr = m.StatusSuccess
		case installer.StatusFailed:
			stStr = m.StatusError
		}

		line := fmt.Sprintf(" %-40s %s", step.Title, stStr)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.SubtitleStyle.Render("Terminal Output Logs:"))
	b.WriteString("\n")
	b.WriteString(m.BoxStyle.Render(m.Viewport.View()))
	b.WriteString("\n")

	if m.State == StateDone {
		if m.Err != nil {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.AppConfig.Theme.StatusError)).Render("\nInstallation failed: "))
			b.WriteString(m.Err.Error())
			b.WriteString("\nPress [ENTER] to exit.\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.AppConfig.Theme.StatusSuccess)).Render("\n✨ Setup and deployment completed successfully!"))
			b.WriteString("\nPress [ENTER] to finish.\n")
		}
	}

	return b.String()
}

func (m *Model) AddLog(logLine string) tea.Cmd {
	return func() tea.Msg {
		return LogMsg(logLine)
	}
}

func (m *Model) SendLog(p *tea.Program, logLine string) {
	if p != nil {
		p.Send(LogMsg(fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), logLine)))
	}
}

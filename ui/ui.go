package ui

import (
	"azdo-dash/config"
	"azdo-dash/context"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"os"
	"strings"
	"time"
)

type ErrMsg error

type Model struct {
	items      []string
	quitting   bool
	err        error
	configPath string
	ctx        context.ProgramContext
}

type initMsg struct {
	Config config.Config
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func NewModel(configPath string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := Model{
		configPath: configPath,
	}
	m.ctx = context.ProgramContext{
		ConfigPath: configPath,
	}

	return m
}

func (m *Model) initScreen() tea.Msg {
	showError := func(err error) {
		styles := log.DefaultStyles()
		styles.Key = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)
		styles.Separator = lipgloss.NewStyle()

		logger := log.New(os.Stderr)
		logger.SetStyles(styles)
		logger.SetTimeFormat(time.RFC3339)
		logger.SetReportTimestamp(true)
		logger.SetPrefix("Reading config file")
		logger.SetReportCaller(true)

		logger.
			Fatal(
				"failed parsing config file",
				"location",
				m.configPath,
				"err",
				err,
			)
	}

	cfg, err := config.ParseConfig(m.ctx.ConfigPath)
	if err != nil {
		showError(err)
		return initMsg{Config: cfg}
	}

	return initMsg{Config: cfg}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initScreen, tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case initMsg:
		m.ctx.Config = &msg.Config

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit

		}
		return m, nil
	case ErrMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		return m, cmd
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.ctx.Config == nil {
		return "Reading config...\n"
	}
	if m.err != nil {
		return m.err.Error()
	}
	if m.quitting {
		return ""
	}

	str := strings.Builder{}
	str.WriteString("Project Ids:\n\n")
	for idx, item := range m.ctx.Config.ProjectIds {
		str.WriteString(fmt.Sprintf("%d. %s\n", idx+1, item))
	}
	str.WriteString("\n\nRepo Ids:\n\n")
	for idx, item := range m.ctx.Config.RepoIds {
		str.WriteString(fmt.Sprintf("%d. %s\n", idx+1, item))
	}

	return str.String()
}

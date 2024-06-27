package ui

import (
	"azdo-dash/config"
	"azdo-dash/constants"
	"azdo-dash/context"
	"azdo-dash/ui/prssection"
	"azdo-dash/ui/section"
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

type initMsg struct {
	Config config.Config
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

type Model struct {
	items       []string
	quitting    bool
	err         error
	configPath  string
	ctx         context.ProgramContext
	prSection   section.Section
	tasks       map[string]context.Task
	taskSpinner spinner.Model
}

func NewModel(configPath string) Model {
	taskSpinner := spinner.Model{Spinner: spinner.Dot}

	m := Model{
		configPath:  configPath,
		tasks:       map[string]context.Task{},
		taskSpinner: taskSpinner,
	}
	m.ctx = context.ProgramContext{
		ConfigPath: configPath,
		StartTask: func(task context.Task) tea.Cmd {
			log.Debug("Starting task", "id", task.Id)
			task.StartTime = time.Now()
			m.tasks[task.Id] = task
			return m.taskSpinner.Tick
		},
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
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case initMsg:
		m.ctx.Config = &msg.Config

		prSection, fetchSectionsCmd := m.fetchAllViewSections()
		m.setCurrentViewSections(prSection)
		cmds = append(cmds, fetchSectionsCmd)

	case constants.TaskFinishedMsg:
		task, ok := m.tasks[msg.TaskId]
		if ok {
			log.Debug("Task finished", "id", task.Id)
			if msg.Err != nil {
				log.Error("Task finished with error", "id", task.Id, "err", msg.Err)
				task.State = context.TaskError
				task.Error = msg.Err
			} else {
				task.State = context.TaskFinished
			}
			now := time.Now()
			task.FinishedTime = &now
			m.tasks[msg.TaskId] = task
			cmd = tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return constants.ClearTaskMsg{TaskId: msg.TaskId}
			})

			m.updateSection(msg.SectionId, msg.SectionType, msg.Msg)
		}

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

	sectionCmd := m.updateCurrentSection(msg)
	cmds = append(cmds, cmd, sectionCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.ctx.Config == nil {
		return "Reading config...\n"
	}

	s := strings.Builder{}
	s.WriteString("\n")
	currSection := m.getCurrSection()
	mainContent := ""
	if currSection != nil {
		mainContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.getCurrSection().View(),
		)
	} else {
		mainContent = "No sections defined..."
	}
	s.WriteString(mainContent)
	s.WriteString("\n")

	return s.String()
}

func (m *Model) setCurrentViewSections(newSections []section.Section) {
	m.prSection = newSections[0]
}

func (m *Model) updateCurrentSection(msg tea.Msg) (cmd tea.Cmd) {
	section := m.getCurrSection()
	if section == nil {
		return nil
	}
	return m.updateSection(section.GetId(), section.GetType(), msg)
}

func (m *Model) fetchAllViewSections() ([]section.Section, tea.Cmd) {
	return prssection.FetchAllSections(m.ctx)
}

func (m *Model) updateSection(id int, sType string, msg tea.Msg) (cmd tea.Cmd) {
	var updatedSection section.Section
	updatedSection, cmd = m.prSection.Update(msg)
	m.prSection = updatedSection

	return cmd
}

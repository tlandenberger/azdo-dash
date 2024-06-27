package context

import (
	"azdo-dash/config"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type ProgramContext struct {
	Config     *config.Config
	ConfigPath string
	StartTask  func(task Task) tea.Cmd
}

type State = int

const (
	TaskStart State = iota
	TaskFinished
	TaskError
)

type Task struct {
	Id           string
	StartText    string
	FinishedText string
	State        State
	Error        error
	StartTime    time.Time
	FinishedTime *time.Time
}

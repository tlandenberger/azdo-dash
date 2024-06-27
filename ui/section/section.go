package section

import (
	"azdo-dash/context"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type Model struct {
	Id                        int
	Ctx                       *context.ProgramContext
	Spinner                   spinner.Model
	IsSearching               bool
	SearchValue               string
	Table                     table.Model
	Type                      string
	SingularForm              string
	PluralForm                string
	Columns                   []table.Column
	TotalCount                int
	IsPromptConfirmationShown bool
	PromptConfirmationAction  string
	LastFetchTaskId           string
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	sType string,
	lastUpdated time.Time,
) Model {
	m := Model{
		Id:      id,
		Type:    sType,
		Ctx:     ctx,
		Spinner: spinner.Model{Spinner: spinner.Dot},
	}

	return m
}

type Section interface {
	Identifier
	Component
}

type Identifier interface {
	GetId() int
	GetType() string
}

type Component interface {
	Update(msg tea.Msg) (Section, tea.Cmd)
	View() string
}

func (m *Model) GetId() int {
	return m.Id
}

func (m *Model) GetType() string {
	return m.Type
}

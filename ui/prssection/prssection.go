package prssection

import (
	"azdo-dash/constants"
	"azdo-dash/context"
	"azdo-dash/data"
	"azdo-dash/ui/section"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
	"strings"
	"time"
)

const SectionType = "pr"

type SectionPullRequestsFetchedMsg struct {
	Prs        []data.PullRequestData
	TotalCount int
	TaskId     string
}

type Model struct {
	section.Model
	Prs        []data.PullRequestData
	TotalCount int
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	lastUpdated time.Time,
) Model {
	m := Model{}
	m.Model = section.NewModel(
		id,
		ctx,
		SectionType,
		lastUpdated,
	)
	m.Prs = []data.PullRequestData{}

	return m
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case SectionPullRequestsFetchedMsg:
		if m.LastFetchTaskId == msg.TaskId {
			m.Prs = msg.Prs

			m.TotalCount = msg.TotalCount
		}
	}

	return &m, tea.Batch(cmd)
}

func (m Model) View() string {
	s := strings.Builder{}

	s.WriteString("PR LIST:")
	for _, pr := range m.Prs {

		required := false
		for _, reviewer := range pr.RequiredReviewers {
			if reviewer == "Tobias Landenberger" {
				required = true
			}
		}

		s.WriteString("\n" + strconv.Itoa(pr.ID) + "\t" + pr.Status + "\t" + strconv.FormatBool(required) + "\t" + pr.RepositoryName)
	}

	return s.String()
}

func FetchAllSections(
	ctx context.ProgramContext,
) (sections []section.Section, fetchAllCmd tea.Cmd) {
	fetchPRsCmds := make([]tea.Cmd, 0)
	sections = make([]section.Section, 0)
	sectionModel := NewModel(
		1,
		&ctx,
		time.Now(),
	) // 0 is the search section
	sections = append(sections, &sectionModel)
	fetchPRsCmds = append(
		fetchPRsCmds,
		sectionModel.FetchNextPageSectionRows()...)
	return sections, tea.Batch(fetchPRsCmds...)
}

func (m *Model) FetchNextPageSectionRows() []tea.Cmd {
	if m == nil {
		return nil
	}

	var cmds []tea.Cmd

	var requests []data.FetchPRRequest
	for _, project := range m.Ctx.Config.Projects {
		for _, repoId := range project.RepoIds {
			requests = append(requests, data.FetchPRRequest{
				OrgName:             m.Ctx.Config.OrgName,
				ProjectID:           project.Id,
				RepoID:              repoId,
				PersonalAccessToken: m.Ctx.Config.PersonalAccessToken,
			})
		}
	}

	startCursor := time.Now().String()
	id := 1
	taskId := fmt.Sprintf("fetching_prs_%d_%s", id, startCursor)
	m.LastFetchTaskId = taskId
	task := context.Task{
		Id:        taskId,
		StartText: fmt.Sprintf(`Fetching PRs for "%s"`, m.Ctx.Config.OrgName),
		FinishedText: fmt.Sprintf(
			`PRs for "%s" have been fetched`,
			m.Ctx.Config.OrgName,
		),
		State: context.TaskStart,
		Error: nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	fetchCmd := func() tea.Msg {
		res, err := data.FetchPullRequests(requests)
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   id,
				SectionType: SectionType,
				TaskId:      taskId,
				Err:         err,
			}
		}

		return constants.TaskFinishedMsg{
			SectionId:   id,
			SectionType: SectionType,
			TaskId:      taskId,
			Msg: SectionPullRequestsFetchedMsg{
				Prs:        res.Prs,
				TotalCount: res.TotalCount,
				TaskId:     taskId,
			},
		}
	}
	cmds = append(cmds, fetchCmd)

	return cmds
}

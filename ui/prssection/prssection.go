package prssection

import (
	"azdo-dash/constants"
	"azdo-dash/context"
	"azdo-dash/data"
	"azdo-dash/ui/section"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

var (
	headerStyle             = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	rowStyle                = lipgloss.NewStyle()
	columnStyle             = lipgloss.NewStyle().Width(22).PaddingRight(2)
	smallColumnStyle        = lipgloss.NewStyle().Width(10).PaddingRight(2).MaxWidth(10).AlignHorizontal(lipgloss.Center)
	statusActive            = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("●")
	statusCompleted         = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("●")
	statusDraft             = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("●")
	checkMark               = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓")
	crossMark               = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗")
	noVote                  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("")
	approved                = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓")
	approvedWithSuggestions = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("✓✍")
	waitForAuthor           = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render("⌛")
	rejected                = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗")
)

func truncateString(str string, maxWidth int) string {
	if len(str) > maxWidth {
		if maxWidth > 3 {
			return str[:maxWidth-3] + "..."
		}
		return str[:maxWidth]
	}
	return str
}

func formatStatus(status string) string {
	switch status {
	case "active":
		return statusActive
	case "completed":
		return statusCompleted
	case "draft":
		return statusDraft
	default:
		return status
	}
}

func formatVote(vote int) string {
	switch vote {
	case 0:
		return noVote
	case 10:
		return approved
	case 5:
		return approvedWithSuggestions
	case -5:
		return waitForAuthor
	case -10:
		return rejected
	default:
		return noVote
	}
}

func formatBool(value bool) string {
	if value {
		return checkMark
	}
	return crossMark
}

func removePrefix(refName string) string {
	return strings.TrimPrefix(refName, "refs/heads/")
}

func (m Model) View() string {
	s := strings.Builder{}

	headers := []string{"Repository", "Title", "CreatedBy", "Status", "Required", "Vote", "SourceBranch", "IsDraft"}
	for _, header := range headers {
		switch header {
		case "Status", "IsDraft", "Vote", "Required":
			s.WriteString(headerStyle.Render(smallColumnStyle.Render(header)))
		default:
			s.WriteString(headerStyle.Render(columnStyle.Render(header)))
		}
	}
	s.WriteString("\n")

	for _, pr := range m.Prs {

		row := []string{
			pr.RepositoryName,
			pr.Title,
			pr.CreatedBy,
			formatStatus(pr.Status),
			formatBool(pr.IsRequiredReviewer),
			formatVote(pr.Vote),
			removePrefix(pr.SourceBranch),
			formatBool(pr.IsDraft),
		}

		for i, col := range row {
			switch headers[i] {
			case "Status", "Required", "Vote", "IsDraft":
				s.WriteString(rowStyle.Render(smallColumnStyle.Render(col)))
			default:
				s.WriteString(rowStyle.Render(columnStyle.Render(truncateString(col, 20))))
			}
		}
		s.WriteString("\n")
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

package tui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/destinyamba/gh-pullkin/internal/pipeline"
	"github.com/destinyamba/gh-pullkin/internal/rank"
)

var (
	pink  = lipgloss.Color("205")
	cyan  = lipgloss.Color("86")
	green = lipgloss.Color("78")
	gray  = lipgloss.Color("245")

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(pink)
	cursorStyle   = lipgloss.NewStyle().Foreground(pink).Bold(true)
	repoStyle     = lipgloss.NewStyle().Foreground(cyan)
	tagStyle      = lipgloss.NewStyle().Foreground(green)
	dimStyle      = lipgloss.NewStyle().Foreground(gray)
	selectedStyle = lipgloss.NewStyle().Bold(true)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

type state int

const (
	loading state = iota
	ready
	failed
)

type LoadFunc func() (pipeline.Result, error)

type loadedMsg struct {
	res pipeline.Result
	err error
}

type openedMsg struct {
	err error
}

type Model struct {
	load    LoadFunc
	state   state
	spinner spinner.Model
	res     pipeline.Result
	err     error
	status  string
	cursor  int
	offset  int
	width   int
	height  int
}

func New(load LoadFunc) Model {
	return Model{
		load:    load,
		spinner: spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(cursorStyle)),
		width:   80,
		height:  24,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		res, err := m.load()
		return loadedMsg{res: res, err: err}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clampOffset()
		return m, nil

	case loadedMsg:
		m.res, m.err = msg.res, msg.err
		m.state = ready
		if msg.err != nil {
			m.state = failed
		}
		return m, nil

	case openedMsg:
		m.status = ""
		if msg.err != nil {
			m.status = "could not open browser: " + msg.err.Error()
		}
		return m, nil

	case spinner.TickMsg:
		if m.state != loading {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.res.Candidates)-1 {
				m.cursor++
			}
		case "enter":
			if c, ok := m.selected(); ok {
				m.status = "opening " + c.Issue.URL
				return m, openURL(c.Issue.URL)
			}
		}
		m.clampOffset()
	}
	return m, nil
}

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = "pullkin"
	return v
}

func (m Model) render() string {
	var b strings.Builder

	switch m.state {
	case loading:
		fmt.Fprintf(&b, "%s\n\n%s Finding issues in your dependencies...\n", titleStyle.Render("pullkin"), m.spinner.View())
		b.WriteString("\n" + dimStyle.Render("q quit"))
		return b.String()

	case failed:
		fmt.Fprintf(&b, "%s\n\n%s\n", titleStyle.Render("pullkin"), errorStyle.Render(m.err.Error()))
		b.WriteString("\n" + dimStyle.Render("q quit"))
		return b.String()
	}

	cs := m.res.Candidates
	switch {
	case m.res.Deps == 0:
		fmt.Fprintf(&b, "%s\n\nNo go.mod or package.json in this folder. Run pullkin where your project's manifest is.\n", titleStyle.Render("pullkin"))
	case len(cs) == 0:
		fmt.Fprintf(&b, "%s\n\nNo fixable issues found in your dependencies.\n", titleStyle.Render("pullkin"))
	default:
		fmt.Fprintf(&b, "%s %s\n\n", titleStyle.Render("pullkin"), dimStyle.Render(fmt.Sprintf("· %d fixable issues", len(cs))))
		end := min(m.offset+m.listHeight(), len(cs))
		for i := m.offset; i < end; i++ {
			b.WriteString(m.row(i, cs[i]) + "\n")
		}
	}

	b.WriteString("\n")
	if m.status != "" {
		b.WriteString(dimStyle.Render(m.status) + "\n")
	}
	if n := len(m.res.Skipped); n > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("%d repos skipped (could not load)", n)) + "\n")
	}
	b.WriteString(dimStyle.Render("↑/↓ move · enter open · q quit"))
	return b.String()
}

func (m Model) row(i int, c rank.Candidate) string {
	pointer := "  "
	if i == m.cursor {
		pointer = cursorStyle.Render("> ")
	}

	repo := fmt.Sprintf("%s#%d", c.Issue.Repo.Name, c.Issue.Number)

	var tag string
	if isGoodFirst(c) {
		tag = " " + tagStyle.Render("good first issue")
	}
	if !c.Dep.Direct {
		tag += " " + dimStyle.Render("indirect")
	}

	used := 2 + len(repo) + 2 + lipgloss.Width(tag)
	title := ansi.Truncate(c.Issue.Title, max(m.width-used, 10), "…")
	if i == m.cursor {
		title = selectedStyle.Render(title)
	}

	return pointer + repoStyle.Render(repo) + "  " + title + tag
}

func (m Model) selected() (rank.Candidate, bool) {
	if m.cursor < 0 || m.cursor >= len(m.res.Candidates) {
		return rank.Candidate{}, false
	}
	return m.res.Candidates[m.cursor], true
}

func (m Model) listHeight() int {
	return max(m.height-6, 1)
}

func (m *Model) clampOffset() {
	h := m.listHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+h {
		m.offset = m.cursor - h + 1
	}
}

func isGoodFirst(c rank.Candidate) bool {
	for _, l := range c.Issue.Labels {
		if strings.EqualFold(l, "good first issue") || strings.EqualFold(l, "good-first-issue") {
			return true
		}
	}
	return false
}

func openURL(url string) tea.Cmd {
	return func() tea.Msg {
		return openedMsg{err: browser.New("", io.Discard, io.Discard).Browse(url)}
	}
}

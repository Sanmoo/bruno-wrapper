package interactive

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sanmoo/bruwrapper/internal/core"
)

type selector struct{}

func NewSelector() core.Selector {
	return &selector{}
}

type item struct {
	title string
	desc  string
}

func (i item) Title() string {
	return i.title
}

func (i item) Description() string {
	return i.desc
}

func (i item) FilterValue() string {
	return i.title
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if sel, ok := m.list.SelectedItem().(item); ok {
				m.choice = sel.title
			}
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

func newListModel(items []list.Item, title string) model {
	delegate := list.NewDefaultDelegate()

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A6E3A1")).
		Bold(true)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A6E3A1"))

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CDD6F4"))

	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BAC2DE"))

	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086"))

	delegate.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086"))

	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89B4FA")).
		Bold(true)

	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CDD6F4"))

	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CDD6F4"))

	l.Help.Styles.ShortKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CDD6F4"))

	l.Help.Styles.ShortDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BAC2DE"))

	l.Help.Styles.ShortSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086"))

	return model{list: l}
}

func (s *selector) SelectCollection(collections []core.Collection) (core.Collection, error) {
	var items []list.Item
	for _, c := range collections {
		items = append(items, item{title: c.Name, desc: c.Path})
	}

	m := newListModel(items, "Select a Collection")
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return core.Collection{}, err
	}

	fm, ok := result.(model)
	if !ok || fm.choice == "" {
		return core.Collection{}, fmt.Errorf("no selection made")
	}

	for _, c := range collections {
		if c.Name == fm.choice {
			return c, nil
		}
	}

	return core.Collection{}, fmt.Errorf("no selection made")
}

func (s *selector) SelectRequest(requests []core.Request) (core.Request, error) {
	var items []list.Item
	for _, r := range requests {
		items = append(items, item{
			title: fmt.Sprintf("%-7s %s", r.Method, r.Name),
			desc:  r.URL,
		})
	}

	m := newListModel(items, "Select a Request")
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return core.Request{}, err
	}

	fm, ok := result.(model)
	if !ok || fm.choice == "" {
		return core.Request{}, fmt.Errorf("no selection made")
	}

	for _, r := range requests {
		display := fmt.Sprintf("%-7s %s", r.Method, r.Name)
		if display == fm.choice {
			return r, nil
		}
	}

	return core.Request{}, fmt.Errorf("no selection made")
}

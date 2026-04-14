package interactive

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sanmoo/bruwrapper/internal/core"
)

type selector struct{}

func NewSelector() core.Selector {
	return &selector{}
}

type item struct {
	title string
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
	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	return model{list: l}
}

func (s *selector) SelectCollection(collections []core.Collection) (core.Collection, error) {
	var items []list.Item
	for _, c := range collections {
		items = append(items, item{title: c.Name})
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
		items = append(items, item{title: fmt.Sprintf("%-7s %s", r.Method, r.Name)})
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

package tabs

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jquag/tui-do/style"
)

type KeyMap struct {
  NextTab key.Binding
  PrevTab key.Binding
}

var DefaultKeyMap = KeyMap{
  NextTab: key.NewBinding(key.WithKeys("]")),
  PrevTab: key.NewBinding(key.WithKeys("[")),
}

type Model struct {
  Tabs []string
  ActiveIndex int
  KeyMap KeyMap
  Width int
}

func New(tabs... string) Model {
  return Model{
    Tabs: tabs,
    ActiveIndex: 0,
    KeyMap: DefaultKeyMap,
  }
}

func (m *Model) next() {
  if m.ActiveIndex + 1 < len(m.Tabs)  {
    m.ActiveIndex++
  }
}

func (m *Model) prev() {
  if m.ActiveIndex - 1 >= 0 {
    m.ActiveIndex--
  }
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
  switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.NextTab):
      m.next()
		case key.Matches(msg, m.KeyMap.PrevTab):
      m.prev()
    }
  }

  return m, nil
}

func (m Model) View() string {
  var tabStrings []string
  for i, t := range m.Tabs {
    if i == m.ActiveIndex {
      tabStrings = append(tabStrings, style.TabActive.Render(t))
    } else {
      tabStrings = append(tabStrings, style.TabInactive.Render(t))
    }
  }
  w := lipgloss.Width(lipgloss.JoinHorizontal(lipgloss.Bottom, tabStrings...))
  if w < m.Width {
    tabStrings = append(tabStrings, style.TabFiller.Render(strings.Repeat(" ", m.Width - w - 4)))
  }
  return lipgloss.JoinHorizontal(lipgloss.Bottom, tabStrings...)
}

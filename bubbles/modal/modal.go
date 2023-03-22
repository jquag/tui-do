package modal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jquag/tui-do/style"
)

type Model struct {
  Width int
  Height int
  Title string
  Body string
  Confirmed bool
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { 
  switch msg := msg.(type) {
    case tea.KeyMsg:
      switch msg.String() {
        case tea.KeyEnter.String(), "y", "Y":
          m.Confirmed = true
          return m, makeCmd(Confirmed) 
        case tea.KeyEscape.String(), "n", "N":
          m.Confirmed = false
          return m, makeCmd(Cancelled) 
      }
  }

  return m, nil
}

func (m Model) View() string {
  title := style.ModalTitle.Render(m.Title)
  body := lipgloss.NewStyle().MaxWidth(m.Width-6).Render(m.Body)
  modal := style.ModalBox.Render(fmt.Sprintf("%s\n%s", title, body))
  return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars("?"), lipgloss.WithWhitespaceForeground(lipgloss.Color("#595959")))
}

type ModalMsg int

const (
  Confirmed ModalMsg = iota
  Cancelled 
)

func makeCmd(msg ModalMsg) tea.Cmd {
  return func() tea.Msg {
    return msg
  }
}

func confirmed() tea.Msg {
  return "confirmed"
}

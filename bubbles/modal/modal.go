package modal

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jquag/tui-do/style"
	"github.com/muesli/reflow/padding"
)

const Marker = '\x1B'

func IsTerminator(c rune) bool {
	return (c >= 0x40 && c <= 0x5a) || (c >= 0x61 && c <= 0x7a)
}

type StyledString struct {
  open string
  content string
  close string
}

type Model struct {
  Width int
  Height int
  Title string
  Body string
  Confirmed bool
  BackgroundView string
}

func ParseStyledString(s string) []StyledString {
  spans := []StyledString{}

  current := StyledString{}
  ansiOpening := false
  ansiClosing := false
  for _, c := range s {
    if c == Marker {
      if current.open == "" {
        if current.content != "" {
          spans = append(spans, current)
          current = StyledString{}
        }
        ansiOpening = true
        current.open += string(c)
      } else {
        ansiClosing = true
        current.close += string(c)
      }
    } else if ansiOpening {
      current.open += string(c)
      if IsTerminator(c) {
        ansiOpening = false
      }
    } else if ansiClosing {
      current.close += string(c)
      if IsTerminator(c) {
        ansiClosing = false
        spans = append(spans, current)
        current = StyledString{}
      }
    } else {
      current.content += string(c)
    }
  }


  if current.close == "" && current.content != "" {
    spans = append(spans, current)
  }

  return spans
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

func ReplaceChunk(s string, startIndex int, replacement string) string {
  if s == "" {
    s = strings.Repeat(" ", startIndex + 1)
  } else {
    s = padding.String(s, uint(startIndex))
  }

  spans := ParseStyledString(s)
  replaced := ""

  replacementWidth := lipgloss.Width(replacement)
  for _, ss := range spans {
    widthSoFar := lipgloss.Width(replaced)
    currentWidth := lipgloss.Width(ss.open + ss.content + ss.close)

    if widthSoFar > startIndex {
      if widthSoFar > startIndex + replacementWidth {
        replaced = replaced + ss.open + ss.content + ss.close
      } else if widthSoFar + currentWidth > startIndex + replacementWidth {
          replaced = replaced + replacement + ss.open + string([]rune(ss.content)[startIndex - widthSoFar + replacementWidth:]) + ss.close
      }
    } else if widthSoFar + currentWidth >= startIndex {
      replaced = replaced + ss.open + string([]rune(ss.content)[0:startIndex - widthSoFar]) + ss.close
      if widthSoFar + currentWidth >= startIndex + replacementWidth {
        replaced = replaced + replacement + ss.open + string([]rune(ss.content)[startIndex - widthSoFar + replacementWidth:]) + ss.close
      }
    } else {
      replaced = replaced + ss.open + ss.content + ss.close
    }
  }

  if lipgloss.Width(replaced) <= startIndex {
    replaced += replacement
  }

  return replaced
}

func (m Model) View() string {
  title := style.ModalTitle.Render(m.Title)
  body := lipgloss.NewStyle().MaxWidth(m.Width-6).Render(m.Body)
  modal := style.ModalBox.Render(fmt.Sprintf("%s\n%s", title, body))

  modalWidth, modalHeight := lipgloss.Size(modal)

  startY := (m.Height/2) - (modalHeight/2)
  startX := (m.Width/2) - (modalWidth/2)

  lines := strings.Split(m.BackgroundView, "\n")
  udpatedLines := []string{}
  modalLines := strings.Split(modal, "\n")

  for i, line := range lines {
    if len(modalLines) > i - startY && i >= startY {
      replaced := ReplaceChunk(line, startX, modalLines[i-startY])
      udpatedLines = append(udpatedLines, replaced)
    } else {
      udpatedLines = append(udpatedLines, line)
    }
  }

  return strings.Join(udpatedLines, "\n")
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

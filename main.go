package main

import (
	"fmt"
	"math"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jquag/tui-do/bubbles/tabs"
	"github.com/jquag/tui-do/repo"
	"github.com/jquag/tui-do/service"
	"github.com/jquag/tui-do/style"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
const useHighPerformanceRenderer = false

type Model struct {
  Svc *service.Service
  IsAdding bool
  IsDeleting bool
  todoCursorRow int
  completedCursorRow int
  Tabs tabs.Model
  ListViewport viewport.Model
  inactiveTabViewportOffset int
  ready bool
} 

func (m Model) cursorRow() int {
  if m.Tabs.ActiveIndex == 0 {
    return m.todoCursorRow
  } else {
    return m.completedCursorRow
  }
}

func (m *Model) incCursorRow() {
  if m.Tabs.ActiveIndex == 0 {
    m.todoCursorRow++
  } else {
    m.completedCursorRow++
  }
}

func (m *Model) decCursorRow() {
  if m.Tabs.ActiveIndex == 0 {
    m.todoCursorRow--
  } else {
    m.completedCursorRow--
  }
}

func initialModel() Model {
  filename := ".tuido.json"
  if (len(os.Args) > 1) {
    filename = os.Args[1]
  }
  r := repo.NewRepo(filename)
  s := service.NewService(r)

  return Model{
    Svc: s,
    Tabs: tabs.New("TODO", "Complete"),
  }
}

func (m Model) Init() tea.Cmd {
  return nil
}

// func write_to_file_cmd(m model) tea.Cmd {
//   return func() tea.Msg {
//     content, _ := json.MarshalIndent(m.todos, "", "  ")
//     os.WriteFile(m.filename, content, 0644)
//     return nil
//   }
// }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  todos := m.Svc.Todos(m.Tabs.ActiveIndex == 1)
  skipViewportUpdate := false
  cursorRow := m.cursorRow()
  var cmds []tea.Cmd

  prevTab := m.Tabs.ActiveIndex
  prevYOffset := m.ListViewport.YOffset
  var cmd tea.Cmd
  m.Tabs, cmd = m.Tabs.Update(msg)
  cmds = append(cmds, cmd)
  tabChanged := prevTab != m.Tabs.ActiveIndex

  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {

    case "ctrl+c", "q":
    return m, tea.Quit

    case "up", "k":
    if cursorRow > 0 {
      m.decCursorRow()
    }
    skipViewportUpdate = cursorRow - m.ListViewport.YOffset >= 2 // b/c cursor is not close to the top

    case "down", "j":
    if cursorRow < len(todos)-1 {
      m.incCursorRow()
    }
    skipViewportUpdate = cursorRow <= m.ListViewport.Height - 3 // b/c cursor is not close to the bottom

    // case "enter", " ":
    //   if m.todos[m.cursor].Done {
    //     m.todos[m.cursor].Done = false
    //   } else {
    //     m.todos[m.cursor].Done = true
    //   }
    //   return m, write_to_file_cmd(m)
  }

  case tea.WindowSizeMsg:
    headerHeight := 5 //TODO: calc this
    footerHeight := 3 //TODO: calc this
    verticalMarginHeight := headerHeight + footerHeight
    if !m.ready {
      m.ListViewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
      m.ListViewport.YPosition = headerHeight
      m.ListViewport.HighPerformanceRendering = useHighPerformanceRenderer
      m.ready = true

      // This is only necessary for high performance rendering, which in
      // most cases you won't need.
      //
      // Render the viewport one line below the header.
      m.ListViewport.YPosition = headerHeight + 1
    } else {
      m.ListViewport.Width = msg.Width
      m.ListViewport.Height = msg.Height - verticalMarginHeight
    }

  // case []todo:
  //   m.todos = msg
  }

  m.ListViewport.SetContent(m.ContentView())

  if tabChanged {
    m.ListViewport.SetYOffset(m.inactiveTabViewportOffset)
    m.inactiveTabViewportOffset = prevYOffset
  }

  if !skipViewportUpdate {
    m.ListViewport, cmd = m.ListViewport.Update(msg)
    cmds = append(cmds, cmd)
  }

  return m, tea.Batch(cmds...)
}

func (m Model) View() string {
  if !m.ready {
		return "\n  Initializing..."
	}

  footer := "\n\n"+style.Muted.Render("Press ? help")
  tabs := m.Tabs.View(int(math.Max(30.0, float64(lipgloss.Width(m.ContentView())))))
  // m.ListViewport.SetContent(m.ContentView())

  return fmt.Sprintf("%s\n\n%s\n%s", tabs, m.ListViewport.View(), footer)
  // return style.Card.Render(tabs + "\n\n" + s)
}

func (m Model) ContentView() string {
  var s string
  for i, todo := range m.Svc.Todos(m.Tabs.ActiveIndex == 1) {
    cursor := " "
    if m.cursorRow() == i {
      cursor = ">"
    }

    checked := " "
    if todo.Done {
      checked = "x"
    }

    if m.cursorRow() == i {
      s += style.Highlight.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, todo.Name))
    } else {
      s += fmt.Sprintf("%s [%s] %s", cursor, checked, todo.Name)
    }
    s += "\n"
  }
  return s
}

func main() {
  p := tea.NewProgram(initialModel(), tea.WithAltScreen())
  if _, err := p.Run(); err != nil {
    fmt.Printf("Alas, there's been an error: %v", err)
    os.Exit(1)
  }
}


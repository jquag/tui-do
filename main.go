package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jquag/tui-do/bubbles/modal"
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
  isAdding bool
  isDeleting bool
  todoCursorRow int
  completedCursorRow int
  Tabs tabs.Model
  ListViewport viewport.Model
  inactiveTabViewportOffset int
  ready bool
  textInput textinput.Model
  width int
  height int
  confirmationModal modal.Model
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

  ti := textinput.New()
	ti.Placeholder = "enter TODO item"
	ti.Width = 20
  ti.Cursor.SetMode(cursor.CursorBlink)

  return Model{
    Svc: s,
    Tabs: tabs.New("TODO", "Complete"),
    textInput: ti,
  }
}

func (m Model) Init() tea.Cmd {
  return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  initialModel := m
  todos := m.Svc.Todos(m.Tabs.ActiveIndex == 1)
  skipViewportUpdate := false
  cursorRow := m.cursorRow()
  var cmds []tea.Cmd

  var cmd tea.Cmd
  m.Tabs, cmd = m.Tabs.Update(msg)
  cmds = append(cmds, cmd)
  tabChanged := initialModel.Tabs.ActiveIndex != m.Tabs.ActiveIndex

  switch msg := msg.(type) {
  case tea.KeyMsg:
    if !m.isAdding && !m.isDeleting {
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

        case "a":
          if m.Tabs.ActiveIndex == 0 {
            m.isAdding = true
            m.textInput.Focus()
            m.textInput.SetValue("")
            cmd := m.textInput.Cursor.BlinkCmd()
            cmds = append(cmds, cmd)
          }

        case tea.KeyEnter.String(), " ":
          cmds = append(cmds, toggleTodoCommand(m.Svc, todos[m.cursorRow()]))

        case "d":
          m.isDeleting = true
          m.confirmationModal.Title = "Are you sure you want to delete the item?"
          m.confirmationModal.Body = todos[m.cursorRow()].Name
      }
    } else if m.isAdding {
      switch msg.String() {
        case "ctrl+c":
          return m, tea.Quit

        case tea.KeyEscape.String():
          m.isAdding = false

        case tea.KeyEnter.String():
          m.isAdding = false
          cmds = append(cmds, addTodoCommand(m.Svc, m.cursorRow(), m.textInput.Value()))
      }
    } else {
      switch msg.String() {
        case "ctrl+c", "q":
          return m, tea.Quit
      }
    }

  case tea.WindowSizeMsg:
    m.Tabs.Width = msg.Width
    m.textInput.Width = msg.Width - 3
    m.width = msg.Width
    m.height = msg.Height
    m.confirmationModal.Width = msg.Width
    m.confirmationModal.Height = msg.Height
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

  case string:
    if msg == "todo-added" {
      m.incCursorRow()
      if (m.cursorRow() >= len(todos)) {
        m.ListViewport.YOffset = m.ListViewport.YOffset + 1
      }
    }

    if msg == "todo-toggled" || msg == "todo-deleted" {
      if (m.cursorRow() >= len(todos)) {
        m.decCursorRow()
      }
    }

  case modal.ModalMsg:
    if msg == modal.Confirmed {
      m.isDeleting = false
      cmds = append(cmds, deleteTodoCommand(m.Svc, todos[m.cursorRow()]))
    } else if msg == modal.Cancelled {
      m.isDeleting = false
    }

  }

  m.ListViewport.SetContent(m.ContentView())

  if tabChanged {
    m.ListViewport.SetYOffset(m.inactiveTabViewportOffset)
    m.inactiveTabViewportOffset = initialModel.ListViewport.YOffset
  }

  if !skipViewportUpdate && !m.isAdding {
    m.ListViewport, cmd = m.ListViewport.Update(msg)
    cmds = append(cmds, cmd)
  }

  if initialModel.isAdding {
    var cmd tea.Cmd
    m.textInput, cmd = m.textInput.Update(msg)
    cmds = append(cmds, cmd)
  }

  if initialModel.isDeleting {
    var cmd tea.Cmd
    m.confirmationModal, cmd = m.confirmationModal.Update(msg)
    cmds = append(cmds, cmd)
  }

  return m, tea.Batch(cmds...)
}

func (m Model) View() string {
  if !m.ready {
		return "\n  Initializing..."
	}

  if m.isDeleting {
    return m.confirmationModal.View()
  }

  footer := "\n\n"+style.Muted.Render("Press ? help")
  // tabs := m.Tabs.View(int(math.Max(30.0, float64(lipgloss.Width(m.ContentView())))))
  tabs := m.Tabs.View()

  return fmt.Sprintf("%s\n\n%s\n%s", tabs, m.ListViewport.View(), footer)
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
      if m.Tabs.ActiveIndex == 0 && m.isAdding {
        s += fmt.Sprintf("%s [%s] %s", " ", checked, todo.Name)
        s += "\n" + m.textInput.View()
      } else {
        s += style.Highlight.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, todo.Name))
      }
    } else {
      s += fmt.Sprintf("%s [%s] %s", cursor, checked, todo.Name)
    }
    s += "\n"
  }
  return s
}

func addTodoCommand(service *service.Service, index int, name string) tea.Cmd {
  return func() tea.Msg {
    service.AddTodo(index, name)
    return "todo-added"
  }
}

func toggleTodoCommand(service *service.Service, item repo.Todo) tea.Cmd {
  return func() tea.Msg {
    service.ToggleTodo(item)
    return "todo-toggled"
  }
}

func deleteTodoCommand(service *service.Service, item repo.Todo) tea.Cmd {
  return func() tea.Msg {
    service.DeleteTodo(item)
    return "todo-deleted"
  }
}

func main() {
  p := tea.NewProgram(initialModel(), tea.WithAltScreen())
  if _, err := p.Run(); err != nil {
    fmt.Printf("Alas, there's been an error: %v", err)
    os.Exit(1)
  }
}


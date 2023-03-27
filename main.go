package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
  isEditing bool
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
	ti.Width = 20
  ti.Cursor.SetMode(cursor.CursorBlink)
  ti.Prompt = ">  "

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
  totalRows := m.countRows(todos)
  skipViewportUpdate := false
  cursorRow := m.cursorRow()
  var cmds []tea.Cmd

  var cmd tea.Cmd
  m.Tabs, cmd = m.Tabs.Update(msg)
  cmds = append(cmds, cmd)
  tabChanged := initialModel.Tabs.ActiveIndex != m.Tabs.ActiveIndex
  currentItem, _ := m.itemAtIndex(todos, m.cursorRow(), 0)

  switch msg := msg.(type) {
  case tea.KeyMsg:
    if !m.isAdding && !m.isDeleting && !m.isEditing {
      switch msg.String() {

        case "ctrl+c", "q":
          return m, tea.Quit

        case "up", "k":
          if cursorRow > 0 {
            m.decCursorRow()
          }
          skipViewportUpdate = cursorRow - m.ListViewport.YOffset >= 2 // b/c cursor is not close to the top

        case "down", "j":
          if cursorRow < totalRows-1 {
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

        case "c":
          m.isEditing = true
          m.textInput.Focus()
          m.textInput.SetValue(currentItem.Name)
          m.textInput.CursorEnd()
          cmd := m.textInput.Cursor.BlinkCmd()
          cmds = append(cmds, cmd)

        case tea.KeyEnter.String(), " ":
          cmds = append(cmds, toggleTodoCommand(m.Svc, *currentItem))

        case "d":
          m.isDeleting = true
          m.confirmationModal.Title = "Are you sure you want to delete the item?"
          m.confirmationModal.Body = currentItem.Name
      }
    } else if m.isAdding || m.isEditing {
      switch msg.String() {
        case "ctrl+c":
          return m, tea.Quit

        case tea.KeyEscape.String():
          m.isAdding = false
          m.isEditing = false

        case tea.KeyEnter.String():
          m.isAdding = false
          m.isEditing = false
          if initialModel.isAdding {
            if len(todos) == 0 {
              cmds = append(cmds, addTodoCommand(m.Svc, nil, m.textInput.Value()))
            } else {
              cmds = append(cmds, addTodoCommand(m.Svc, currentItem, m.textInput.Value()))
            }
          } else {
            cmds = append(cmds, changeTodoCommand(m.Svc, *currentItem, m.textInput.Value()))
          }
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
      if len(todos) > 1 {
        m.incCursorRow()
      }
      if (m.cursorRow() >= totalRows) {
        m.ListViewport.YOffset = m.ListViewport.YOffset + 1
      }
    }

    if msg == "todo-toggled" || msg == "todo-deleted" {
      if (len(todos) > 0 && m.cursorRow() >= totalRows) {
        m.decCursorRow()
      }
    }

  case modal.ModalMsg:
    if msg == modal.Confirmed {
      m.isDeleting = false
      cmds = append(cmds, deleteTodoCommand(m.Svc, *currentItem))
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

  if initialModel.isAdding || initialModel.isEditing {
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

  footer := "\n\n"+style.Muted.Render("Press ? for help")
  tabs := m.Tabs.View()

  return fmt.Sprintf("%s\n\n%s\n%s", tabs, m.ListViewport.View(), footer)
}

func (m Model) ContentView() string {
  var s string
  todos := m.Svc.Todos(m.Tabs.ActiveIndex == 1)

  if !m.isAdding && len(todos) == 0 {
    return style.Muted.Render("No items")
  }

  if m.isAdding && len(todos) == 0 {
    return m.textInput.View()
  }

  visibleChildCount := 0
  for i, todo := range todos {
    itemString, _ := m.ItemView(todo, visibleChildCount + i, "")
    s += itemString
    visibleChildCount += m.countRows(todo.Children)
  }
  return s
}

func (m Model) ItemView(item repo.Todo, index int, padding string) (string, int) {
  var s string

  hasChildren := len(item.Children) > 0
  isCurrentRow := m.cursorRow() == index

  var prefix string
  outerStyle := lipgloss.NewStyle().Inherit(style.Muted)
  innerStyle := lipgloss.NewStyle()
  if isCurrentRow && !m.isAdding {
    outerStyle = style.Muted.Copy().Inherit(style.Highlight)
    innerStyle = style.CheckBox
  }
  if hasChildren {
    prefix = fmt.Sprintf("%s%s%s", outerStyle.Render(" "), innerStyle.Render("─"), outerStyle.Render(" "))
  } else {
    checked := " "
    if item.Done {
      checked = "x"
    }
    prefix = fmt.Sprintf("%s%s%s", outerStyle.Render("["), innerStyle.Render(checked), outerStyle.Render("]"))
  }

  if isCurrentRow {
    if m.Tabs.ActiveIndex == 0 && m.isAdding {
      s += fmt.Sprintf("%s %s %s", padding, prefix, item.Name)
      if !hasChildren {
        s += "\n  " + padding + m.textInput.View()
      }
    } else if m.isEditing {
      s += "  " + padding + m.textInput.View()
    } else {
      prePrefix := style.Highlight.Render(fmt.Sprintf("%s ", padding))
      postPrefix := style.Highlight.Render(fmt.Sprintf(" %s", item.Name))
      s += fmt.Sprintf("%s%s%s", prePrefix, prefix, postPrefix)
    }
  } else {
    s += fmt.Sprintf("%s %s %s", padding, prefix, item.Name)
  }

  s += "\n"

  if hasChildren {
    for _, child := range item.Children {
      var childString string
      childString, index = m.ItemView(child, index + 1, padding + "    ")
      s += childString
    }
  }

  if isCurrentRow && m.Tabs.ActiveIndex == 0 && m.isAdding && hasChildren {
    s += "  " + padding + m.textInput.View() + "\n"
  }

  return s, index
}

func (m Model) countRows(items []repo.Todo) int {
  c := len(items)
  for _, item := range items {
    c += m.countRows(item.Children)
  }
  return c
}

func (m Model) itemAtIndex(items []repo.Todo, index int, startingAt int) (*repo.Todo, int) {
  if len(items) == 0 {
    return nil, startingAt
  }

  i := startingAt
  for _, item := range items {
    if index == i {
      return &item, index 
    }
    found, lastIndexChecked := m.itemAtIndex(item.Children, index, i + 1)
    if (found != nil) {
      return found, index
    }
    i = lastIndexChecked
  }

  return nil, i
}

func addTodoCommand(service *service.Service, afterItem *repo.Todo, name string) tea.Cmd {
  return func() tea.Msg {
    service.AddTodo(afterItem, name)
    return "todo-added"
  }
}

func changeTodoCommand(service *service.Service, item repo.Todo, name string) tea.Cmd {
  return func() tea.Msg {
    service.ChangeTodo(item, name)
    return "todo-changed"
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


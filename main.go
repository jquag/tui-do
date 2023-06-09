package main

import (
	"fmt"
	"os"
	"strings"

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
  isAddingChild bool
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
  helpModal modal.Model
  isShowingHelp bool
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

func (m *Model) setCursorRow(row int) {
  if m.Tabs.ActiveIndex == 0 {
    m.todoCursorRow = row
  } else {
    m.completedCursorRow = row
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
  ti.TextStyle = style.ActionStyle
  ti.PromptStyle = ti.PromptStyle.Inherit(style.ActionStyle)

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
    if !m.isAdding && !m.isAddingChild && !m.isDeleting && !m.isEditing && !m.isShowingHelp {
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

        case "A":
          if m.Tabs.ActiveIndex == 0 {
            m.isAddingChild = true
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
          if currentItem != nil {
            if len(currentItem.Children) > 0 {
              cmds = append(cmds, toggleExpandedCommand(m.Svc, *currentItem))
            } else {
              cmds = append(cmds, toggleTodoCommand(m.Svc, *currentItem))
            }
          }

        case "d":
          m.isDeleting = true
          m.confirmationModal.Title = "Are you sure you want to delete the item?"
          m.confirmationModal.Body = currentItem.Name + "\n\n" + style.Muted.Render("ENTER-yes, ESC-no")

        case "?":
          m.isShowingHelp = true
          m.helpModal.Title = "Key Mappings"
          m.helpModal.Body = m.helpBodyView()

        case "G":
          m.setCursorRow(m.countRows(todos) - 1)
          m.ListViewport.SetYOffset(m.ListViewport.Height)

        case "g":
          m.setCursorRow(0)
          m.ListViewport.SetYOffset(0)

        case "W":
          cmds = append(cmds, collapseAllCommand(m.Svc, m.Tabs.ActiveIndex == 1))
      }
    } else if m.isAdding || m.isAddingChild || m.isEditing {
      switch msg.String() {
        case "ctrl+c":
          return m, tea.Quit

        case tea.KeyEscape.String():
          m.isAdding = false
          m.isAddingChild = false
          m.isEditing = false

        case tea.KeyEnter.String():
          m.isAdding = false
          m.isAddingChild = false
          m.isEditing = false
          if initialModel.isAdding {
            if len(todos) == 0 {
              cmds = append(cmds, addTodoCommand(m.Svc, nil, m.textInput.Value()))
            } else {
              cmds = append(cmds, addTodoCommand(m.Svc, currentItem, m.textInput.Value()))
            }
          } else if initialModel.isEditing {
            cmds = append(cmds, changeTodoCommand(m.Svc, *currentItem, m.textInput.Value()))
          } else if initialModel.isAddingChild {
              cmds = append(cmds, addTodoAsChildCommand(m.Svc, currentItem, m.textInput.Value()))
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
    m.helpModal.Width = msg.Width
    m.helpModal.Height = msg.Height
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
    if msg == "todo-added" || msg == "todo-child-added" {
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
      if m.isDeleting {
        m.isDeleting = false
        cmds = append(cmds, deleteTodoCommand(m.Svc, *currentItem))
      }
    } else if msg == modal.Cancelled {
      m.isDeleting = false
      m.isShowingHelp = false
    }

  }

  m.ListViewport.SetContent(m.ContentView())

  if tabChanged {
    m.ListViewport.SetYOffset(m.inactiveTabViewportOffset)
    m.inactiveTabViewportOffset = initialModel.ListViewport.YOffset
  }

  if !skipViewportUpdate && !m.isAdding && !m.isAddingChild && !m.isEditing {
    m.ListViewport, cmd = m.ListViewport.Update(msg)
    cmds = append(cmds, cmd)
  }

  if initialModel.isAdding || initialModel.isAddingChild || initialModel.isEditing {
    var cmd tea.Cmd
    m.textInput, cmd = m.textInput.Update(msg)
    cmds = append(cmds, cmd)
  }

  if initialModel.isDeleting {
    var cmd tea.Cmd
    m.confirmationModal, cmd = m.confirmationModal.Update(msg)
    cmds = append(cmds, cmd)
  }

  if initialModel.isShowingHelp {
    var cmd tea.Cmd
    m.helpModal, cmd = m.helpModal.Update(msg)
    cmds = append(cmds, cmd)
  }

  return m, tea.Batch(cmds...)
}

func (m Model) View() string {
  if !m.ready {
		return "\n  Initializing..."
	}

  footer := "\n\n"+style.Muted.Render("Press ? for help")
  tabs := m.Tabs.View()

  content := fmt.Sprintf("%s\n\n%s\n%s", tabs, m.ListViewport.View(), footer)

  if m.isDeleting {
    m.confirmationModal.BackgroundView = content
    return m.confirmationModal.View()
  } else if m.isShowingHelp {
    m.helpModal.BackgroundView = content
    return m.helpModal.View()
  }

  return content
}

func (m Model) ContentView() string {
  var s string
  todos := m.Svc.Todos(m.Tabs.ActiveIndex == 1)

  if !m.isAdding && len(todos) == 0 {
    return style.Muted.Render(" No items")
  }

  if m.isAdding && len(todos) == 0 {
    return " " + m.textInput.View()
  }

  index := 0
  for _, todo := range todos {
    var itemString string
    itemString, index = m.ItemView(todo, index, "")
    index++
    s += itemString
  }
  return s
}

func (m Model) ItemView(item repo.Todo, index int, padding string) (string, int) {
  var s string

  hasChildren := len(item.Children) > 0
  isCurrentRow := m.cursorRow() == index

  var prefix string
  outerStyle := lipgloss.NewStyle().Inherit(style.CheckBoxBracket)
  innerStyle := lipgloss.NewStyle().Inherit(style.ActionStyle)
  nameStyle := lipgloss.NewStyle().Bold(false)
  if item.Done {
    nameStyle.Inherit(style.Muted)
  }
  if isCurrentRow && !m.isAdding && !m.isAddingChild {
    outerStyle = style.CheckBoxBracket.Copy().Inherit(style.Highlight)
    innerStyle = style.CheckBox.Copy()
  }
  if hasChildren || (m.isAddingChild && isCurrentRow) {
    nameStyle.Inherit(style.ParentColor)
    symbol := "+"
    if item.Expanded || (m.isAddingChild && isCurrentRow) {
      symbol = "-"
    }
    prefix = fmt.Sprintf("%s%s%s", outerStyle.Render("("), innerStyle.Render(symbol), outerStyle.Render(")"))
  } else {
    checked := " "
    if item.Done {
      checked = "x"
    }
    prefix = fmt.Sprintf("%s%s%s", outerStyle.Render("["), innerStyle.Render(checked), outerStyle.Render("]"))
  }

  if isCurrentRow {
    if m.Tabs.ActiveIndex == 0 && m.isAdding {
      s += fmt.Sprintf("%s %s %s", padding, prefix, nameStyle.Render(item.Name))
      if !hasChildren {
        s += "\n  " + padding + m.textInput.View()
      }
    } else if m.isEditing {
      s += "  " + padding + m.textInput.View()
    } else if m.isAddingChild {
      s += fmt.Sprintf("%s %s %s", padding, prefix, nameStyle.Render(item.Name))
      s += "\n  " + padding + "   " + m.textInput.View()
    } else {
      prePrefix := style.Highlight.Render(fmt.Sprintf("%s ", padding))
      postPrefix := style.Highlight.Render(fmt.Sprintf(" %s", nameStyle.Render(item.Name)))
      s += fmt.Sprintf("%s%s%s", prePrefix, prefix, postPrefix)
    }
  } else {
    s += fmt.Sprintf("%s %s %s", padding, prefix, nameStyle.Render(item.Name))
  }

  s += "\n"

  if hasChildren && item.Expanded {
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

func (m Model) helpBodyView() string {
  lines := []string{}
  if (m.Tabs.ActiveIndex == 0) {
    lines = append(lines, "a      " + style.ActionStyle.Render("add new item"))
    lines = append(lines, "A      " + style.ActionStyle.Render("add new item as child"))
  }

  lines = append(lines, "c      " + style.ActionStyle.Render("change item"))
  lines = append(lines, "d      " + style.ActionStyle.Render("delete item"))
  lines = append(lines, "space  " + style.ActionStyle.Render("toggle item"))
  lines = append(lines, "j      " + style.ActionStyle.Render("move down"))
  lines = append(lines, "k      " + style.ActionStyle.Render("move up"))
  lines = append(lines, "W      " + style.ActionStyle.Render("collapse all"))
  lines = append(lines, "G      " + style.ActionStyle.Render("go to bottom"))
  lines = append(lines, "g      " + style.ActionStyle.Render("go to top"))
  lines = append(lines, "]      " + style.ActionStyle.Render("next tab"))
  lines = append(lines, "[      " + style.ActionStyle.Render("prev tab"))
  lines = append(lines, "q      " + style.ActionStyle.Render("quit"))

  return "\n" + strings.Join(lines, "\n") + "\n\n"  + style.Muted.Render("ESC-close")
}

func (m Model) countRows(items []repo.Todo) int {
  c := len(items)
  for _, item := range items {
    if item.Expanded {
      c += m.countRows(item.Children)
    }
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
    if item.Expanded {
      found, lastIndexChecked := m.itemAtIndex(item.Children, index, i + 1)
      if (found != nil) {
        return found, index
      }
      i = lastIndexChecked
    } else {
      i++
    }
  }

  return nil, i
}

func addTodoCommand(service *service.Service, afterItem *repo.Todo, name string) tea.Cmd {
  return func() tea.Msg {
    service.AddTodo(afterItem, name)
    return "todo-added"
  }
}

func addTodoAsChildCommand(service *service.Service, parent *repo.Todo, name string) tea.Cmd {
  return func() tea.Msg {
    service.AddTodoAsChild(parent, name)
    return "todo-child-added"
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

func toggleExpandedCommand(service *service.Service, item repo.Todo) tea.Cmd {
  return func() tea.Msg {
    service.ToggleExpanded(item)
    return "todo-expand-toggled"
  }
}

func deleteTodoCommand(service *service.Service, item repo.Todo) tea.Cmd {
  return func() tea.Msg {
    service.DeleteTodo(item)
    return "todo-deleted"
  }
}

func collapseAllCommand(service *service.Service, completed bool) tea.Cmd {
  return func() tea.Msg {
    service.CollapseAll(completed)
    return "todos-collapsed"
  }
}

func main() {
  p := tea.NewProgram(initialModel(), tea.WithAltScreen())
  if _, err := p.Run(); err != nil {
    fmt.Printf("Alas, there's been an error: %v", err)
    os.Exit(1)
  }
}


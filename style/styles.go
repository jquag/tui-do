package style

import "github.com/charmbracelet/lipgloss"

var tabActiveBorder = lipgloss.Border{
  Top: "─",
  Bottom: " ",
  Left: "│",
  Right: "│",
  TopRight: "┐",
  TopLeft: "┌",
  BottomRight: "└",
  BottomLeft: "┘",
}

var tabInactiveBorder = lipgloss.Border{
  Top: "─",
  Bottom: "─",
  Left: "│",
  Right: "│",
  TopRight: "┐",
  TopLeft: "┌",
  BottomRight: "┴",
  BottomLeft: "┴",
}

var tabFillerBorder = lipgloss.Border{
  Top: " ",
  Bottom: "─",
  Left: " ",
  Right: " ",
  TopRight:  " ",
  TopLeft: " ",
  BottomRight: "─",
  BottomLeft: "─",
}

var Highlight = lipgloss.NewStyle().
  // Background(lipgloss.Color("#af8c8c")).
  Background(lipgloss.Color("#353535"))
  // Foreground(lipgloss.Color("#060606"))
  // Foreground(lipgloss.Color("#060606"))


var Card = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.NormalBorder(), false)

var TabActive = lipgloss.NewStyle().
  Bold(true).
  Border(tabActiveBorder, true).
  Padding(0, 1).
  Foreground(lipgloss.Color("#87a987")).
  BorderForeground(lipgloss.Color("#595959"))

var TabInactive = lipgloss.NewStyle().
  Border(tabInactiveBorder, true).
  Padding(0, 1).
  BorderForeground(lipgloss.Color("#595959")).
  Inherit(Muted)

var TabFiller = lipgloss.NewStyle().
  Border(tabFillerBorder, true).
  Padding(0, 1).
  BorderForeground(lipgloss.Color("#595959"))

var Muted = lipgloss.NewStyle().Foreground(lipgloss.Color("#595959"))

var ModalBox = lipgloss.NewStyle().Padding(0, 2).Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("#00ff00"))
var ModalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#87a987")).Bold(true)

var CheckBox = lipgloss.NewStyle().Background(lipgloss.Color("#ffcbcd")).Foreground(lipgloss.Color("#151837"))

var CheckBoxBracket = lipgloss.NewStyle().Foreground(lipgloss.Color("#deae81"))
var ActionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00b1ff"))
var ParentColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#87a987"))

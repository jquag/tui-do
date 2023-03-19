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
  Bold(false).
  Background(lipgloss.Color("#2d383f"))

var Title = lipgloss.NewStyle().
  Bold(true).
  Foreground(lipgloss.Color("#a0c278")).
  Border(lipgloss.NormalBorder(), false, true, false, false)

var Card = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.NormalBorder(), false)

var TabActive = lipgloss.NewStyle().
  Bold(true).
  Border(tabActiveBorder, true).
  Padding(0, 1).
  Foreground(lipgloss.Color("#a0c278")).
  BorderForeground(lipgloss.Color("#595959"))

var TabInactive = lipgloss.NewStyle().
  Border(tabInactiveBorder, true).
  Padding(0, 1).
  BorderForeground(lipgloss.Color("#595959"))

var TabFiller = lipgloss.NewStyle().
  Border(tabFillerBorder, true).
  Padding(0, 1).
  BorderForeground(lipgloss.Color("#595959"))

var Muted = lipgloss.NewStyle().Foreground(lipgloss.Color("#595959"))


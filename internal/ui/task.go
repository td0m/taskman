package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	TaskIcon     = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	TaskTitle    = lipgloss.NewStyle().Bold(true)
	SubTaskTitle = lipgloss.NewStyle().Foreground(Secondary)
)
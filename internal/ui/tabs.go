package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	tabContainer = lipgloss.NewStyle().Padding(1, 1)
	activeTab    = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	inactiveTab  = lipgloss.NewStyle().Foreground(Secondary)
	tabDivider   = lipgloss.NewStyle().Foreground(Faded)
)

type Tabs struct {
	tabs []string
	i    int

	Width int
	Info  string
}

// NewTabs creates a new tabs ui bubbletea model
func NewTabs(tabs []string) Tabs {
	return Tabs{tabs: tabs}
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m Tabs) Init() tea.Cmd {
	m.Set(0)
	return nil
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m Tabs) Update(_ tea.Msg) (Tabs, tea.Cmd) {
	// panic("not implemented") // TODO: Implement
	return m, nil
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m Tabs) View() string {
	tabs := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		r := inactiveTab
		if i == m.i {
			r = activeTab
		}
		tabs[i] = r.Render(t)
	}
	w := lipgloss.Width
	left := strings.Join(tabs, tabDivider.Render(" | "))
	right := m.Info
	space := lipgloss.NewStyle().Width(m.Width - 2 - w(left) - w(right)).Render("")
	return tabContainer.Render(lipgloss.JoinHorizontal(lipgloss.Center, left, space, right)) + "\n"
}

func (m Tabs) Value() int {
	return m.i
}

func (m *Tabs) Set(i int) {
	m.i = min(max(i, 0), len(m.tabs)-1)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

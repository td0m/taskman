package main

import (
	"strconv"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/td0m/taskman/task"
)

var (
	headerHeight = 3
	footerHeight = 1
)

type path []task.ID

type app struct {
	viewport viewport.Model

	all task.Tasks

	visible []path
	cursor  int
}

// newApp creates a new taskman TUI app
func newApp() app {
	return app{
		all:      task.Tasks{"0": {}},
		viewport: viewport.Model{},
	}
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m app) Init() tea.Cmd {
	return nil
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		// cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		verticalMargins := headerHeight + footerHeight
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMargins
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch msg.String() {
		case "j":
			m.setCursor(m.cursor + 1)
		case "k":
			m.setCursor(m.cursor - 1)
		case "o":
			parent, id := info(m.atCursor())
			err := m.all.Add(parent, id, 1)
			if err != nil {
				panic(err)
			}
			m.setCursor(m.cursor + 1)
			m.updateVisible()
		}
	}

	m.viewport.SetContent(m.renderTasks())

	return m, tea.Batch(cmds...)
}

func (m *app) setCursor(value int) {
	size := len(m.visible)
	m.cursor = clamp(value, 0, max(size-1, 0))
	// update viewport
	if m.cursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	}
	if m.cursor < m.viewport.YOffset {
		m.viewport.YOffset = m.cursor
	}
}

func (m *app) updateVisible() {
	m.visible = traverse(m.all, "0")[1:]
	// TODO: clamp cursor
}

func (m app) atCursor() path {
	// if no items visible
	if m.cursor >= len(m.visible) {
		return []task.ID{}
	}
	return m.visible[m.cursor]
}

func info(path path) (task.ID, task.ID) {
	return getParent(path), getID(path)
}

func getID(path path) task.ID {
	if len(path) < 1 {
		return ""
	}
	return path[len(path)-1]
}

func getParent(path path) task.ID {
	// no parent -> assume root
	if len(path) < 2 {
		return "0"
	}
	return path[len(path)-2]
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m app) View() string {
	return m.viewport.View()
}

func (m app) renderTasks() string {
	s := ""
	for i := range m.visible {
		if i == m.cursor {
			s += "> "
		}
		s += strconv.Itoa(i) + "line\n"
	}
	return s
}

func traverse(m task.Tasks, id task.ID) []path {
	v := m[id]
	all := []path{{id}}
	for _, child := range v.Children {
		childPaths := traverse(m, child)
		for _, subp := range childPaths {
			path := append([]task.ID{id}, subp...)
			all = append(all, path)
		}
	}
	return all
}
func clamp(v, low, high int) int {
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

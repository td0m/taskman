package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/td0m/taskman/pkg/dateinput"
	"github.com/td0m/taskman/task"
	"github.com/td0m/taskman/ui"
)

var (
	headerHeight = 3
	footerHeight = 1
)

type mode int

const (
	normalMode mode = iota
	titleMode
	dateMode
)

type path []task.ID

type app struct {
	viewport  viewport.Model
	dateinput dateinput.Model

	mode mode

	all task.Tasks

	visible []path
	cursor  int
}

// newApp creates a new taskman TUI app
func newApp() app {
	return app{
		all:       task.Tasks{"0": {}},
		viewport:  viewport.Model{},
		dateinput: dateinput.NewModel(),
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
		cmd  tea.Cmd
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
		if msg.Type == tea.KeyEsc {
			m.mode = normalMode
		}
		switch m.mode {
		case dateMode:
			if msg.Type == tea.KeyEnter {
				m.mode = normalMode
				id := getID(m.atCursor())
				err := m.all.SetDue(id, m.dateinput.Value())
				if err != nil {
					panic(err)
				}
			} else {
				m.dateinput, cmd = m.dateinput.Update(msg)
				cmds = append(cmds, cmd)
			}
		case normalMode:
			anchor := 1
			switch msg.String() {
			case "d":
				m.dateinput.SetValue(nil)
				m.mode = dateMode
			case "j":
				m.setCursor(m.cursor + 1)
			case "k":
				m.setCursor(m.cursor - 1)
			case "t":
				id := getID(m.atCursor())
				now := time.Now()
				var err error
				if m.all[id].Done == nil {
					err = m.all.SetDone(id, &now)
				} else {
					err = m.all.SetDone(id, nil)
				}
				if err != nil {
					panic(err)
				}
				m.updateVisible()
			case "H":
				c := m.cursor
				parent, id := info(m.atCursor())
				if m.moveUpLeft() {
					newParent, above := info(m.atCursor())
					m.all.Move(id, parent, newParent, above, 1)
					m.updateVisible()
					m.setCursor(c)
				}
			case "L":
				c := m.cursor
				parent, id := info(m.atCursor())
				if m.moveSameParent(-1) {
					_, above := info(m.atCursor())
					m.all.Move(id, parent, above, "", 1)
					m.updateVisible()
					m.setCursor(c)
				}
			case "O":
				anchor = 0
				fallthrough
			case "o":
				parent, id := info(m.atCursor())
				err := m.all.Add(parent, id, anchor)
				if err != nil {
					panic(err)
				}
				m.updateVisible()
				m.setCursor(m.cursor + anchor)
			}
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

func (m *app) moveSameParent(inc int) bool {
	if len(m.visible) == 0 {
		return false
	}
	path := m.atCursor()
	all := m.visible
	i := m.cursor + inc
	for i >= 0 && i < len(all) {
		p := all[i]
		// prevents from jumping to weird locations
		if len(p) < len(path) {
			return false
		}
		if len(p) == len(path) {
			m.setCursor(i)
			return true
		}
		i += inc
	}
	return false
}

func (m *app) moveUpLeft() bool {
	if len(m.visible) == 0 {
		return false
	}
	path := m.visible[m.cursor]
	before := m.visible[:m.cursor]
	for i := len(before) - 1; i >= 0; i-- {
		p := before[i]
		if len(p) < len(path) {
			m.setCursor(i)
			return true
		}
	}
	return false
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
	statusline := ""
	{
		switch m.mode {
		case dateMode:
			statusline = m.dateinput.View()
		}
	}
	return "\n\n\n" + m.viewport.View() + "\n" + statusline
}

func (m app) renderTasks() string {
	s := ""
	for i, path := range m.visible {
		// s += strconv.Itoa(i) + "line\n"
		task := m.all[getID(path)]
		// prevLen := 0
		// if i > 0 {
		// 	prevLen = len(m.visible[i-1])
		// }
		// nextLen := 0

		// s +=
		if len(path) > 2 {
			s += strings.Repeat("  ", len(path)-2)
		}

		s += ui.RenderIcon(task)
		title := ui.Title(task)
		if i == m.cursor {
			title = title.Copy().Background(ui.Faded).Foreground(ui.Background)
		}
		s += title.Render(task.Title)
		if task.Done == nil {
			s += ui.RenderDue(task)
		}
		s += "\n"
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

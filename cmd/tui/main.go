package main

import (
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/td0m/taskman/internal/ui"
	"github.com/td0m/taskman/pkg/task"
)

var icons = map[string]rune{
	"uni":    '📚',
	"work":   '👔',
	"goals":  '🎯',
	"chores": '🧹',
	"social": '🧑',
	"other":  '💡',
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	a := &app{
		viewport: viewport.Model{},
		tabs:     ui.NewTabs([]string{"Outline", "Today"}),
		store:    task.NewStore(),
	}
	p := tea.NewProgram(a)
	p.EnableMouseAllMotion()
	defer p.DisableMouseAllMotion()
	p.EnterAltScreen()
	defer p.ExitAltScreen()

	check(p.Start())
}

const (
	headerHeight = 3
	footerHeight = 1
)

type mode int

const (
	modeNormal mode = iota
	modeRename
	modeDate
)

type path []task.ID

type app struct {
	mode mode

	viewport viewport.Model
	tabs     ui.Tabs

	cursor  int
	visible []path

	store task.StoreManager
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m app) Init() tea.Cmd {
	return nil
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		verticalMargins := headerHeight + footerHeight
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMargins
		m.tabs.Width = msg.Width
		m.render()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			m.update(msg)
		}
	}
	return m, nil
}

// handle keys differently based on the current mode
func (m *app) update(msg tea.KeyMsg) {
	switch m.mode {
	case modeNormal:
		switch msg.String() {
		case "alt+1":
			m.tabs.Set(0)
		case "alt+2":
			m.tabs.Set(1)
		case "j":
			m.setCursor(m.cursor + 1)
			m.render()
		case "k":
			m.setCursor(m.cursor - 1)
			m.render()
		case "o":
			m.store.Create(task.RandomID(), time.Now())
			m.updateTasks()
		}
	}
}

// updateTasks triggers a rerender of the viewport with all the tasks
func (m *app) updateTasks() {
	m.visible = traverse(m.store, "root")[1:]

	// m.visible = m.filter(m.visible, m.predicates[m.tabs.Value()])
	m.render()
}
func (m *app) render() {
	m.viewport.SetContent(m.viewTasks())
}

func (m *app) setCursor(value int) {
	size := len(m.visible)
	m.cursor = clamp(value, 0, max(size-1, 0))

	// for when no tasks
	if size == 0 {
		return
	}

	tasks := m.visible[:m.cursor]
	linesBeforeCursor := 0
	for i := range tasks {
		linesBeforeCursor += m.sizeOf(i)
	}

	cursorSize := m.sizeOf(m.cursor)

	if linesBeforeCursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.YOffset = linesBeforeCursor + cursorSize - m.viewport.Height
	}

	if linesBeforeCursor <= m.viewport.YOffset {
		m.viewport.YOffset = linesBeforeCursor
	}

}

func (m app) sizeOf(i int) int {
	var (
		currentPath = m.visible[i]
	)
	// rules for separating top level todos
	if len(currentPath) == 2 {
		return 2
	}
	return 1

}

func (m app) viewTasks() string {
	s := ""
	for i, id := range m.store.GetChildren("root") {
		var (
			bigspace bool
			title    lipgloss.Style = ui.TaskTitle
		)
		t := m.store.Get(id)

		// style differences
		if m.store.GetParent(id) == "root" {
			bigspace = true
		} else {
			title = ui.SubTaskTitle
		}
		if i == m.cursor && m.mode == modeNormal {
			title = title.Copy().Background(ui.Faded)
		}

		// renderer
		if bigspace {
			s += "\n"
		}
		s += ui.TaskIcon.Render(string(getIcon(t.Category)))
		s += title.Render(t.Name + "name")
		s += "\n"
	}
	return s
}

func getIcon(category string) rune {
	if category == "" {
		return '∙'
	}
	return icons[category]
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m app) View() string {
	return m.tabs.View() + m.viewport.View() + "\n" + "-"
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

func traverse(s task.StoreManager, id task.ID) []path {
	all := []path{{id}}
	if s.Get(id).Folded {
		return all
	}
	for _, child := range s.GetChildren(id) {
		childPaths := traverse(s, child)
		for _, subp := range childPaths {
			path := append([]task.ID{id}, subp...)
			all = append(all, path)
		}
	}
	return all
}

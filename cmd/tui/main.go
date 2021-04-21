package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
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
	i := textinput.NewModel()
	i.Focus()
	i.Prompt = ""
	i.Width = 40

	a := &app{
		nameinput: i,
		viewport:  viewport.Model{},
		tabs:      ui.NewTabs([]string{"Outline", "Today"}),
		store:     task.NewStore(),
		time:      time.Now(),
		timeSetAt: time.Now(),
	}

	a.predicates = []predicate{
		func(t task.Info) bool { return true },
		func(t task.Info) bool {
			due := t.NextDue()
			return due != nil && due.Before(a.now())
		},
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

type predicate func(task.Info) bool

type app struct {
	mode mode

	viewport  viewport.Model
	nameinput textinput.Model
	tabs      ui.Tabs

	time      time.Time
	timeSetAt time.Time

	cursor     int
	visible    []path
	predicates []predicate

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
	var (
		cmd tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		verticalMargins := headerHeight + footerHeight
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMargins
		m.tabs.Width = msg.Width
		m.setCursor(m.cursor) // make sure cursor is visible
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			m.mode = modeNormal
		default:
			cmd = m.keyUpdate(msg)
		}
	}
	m.render()
	return m, cmd
}

// handle keys differently based on the current mode
func (m *app) keyUpdate(msg tea.KeyMsg) tea.Cmd {
	if m.mode == modeRename || m.mode == modeNormal {
		if msg.Type == tea.KeyTab {
			c := m.cursor
			id := getID(m.atCursor())
			if m.moveSameParent(-1) {
				above := getID(m.atCursor())
				m.store.Move(id, above, task.Into)
				m.updateTasks()
				m.setCursor(c)
			}
			return nil
		}
		if msg.Type == tea.KeyShiftTab {
			c := m.cursor
			id := getID(m.atCursor())
			if m.moveUpLeft() {
				above := getID(m.atCursor())
				m.store.Move(id, above, task.Below)
				m.updateTasks()
				m.setCursor(c)
			}
			return nil
		}
	}
	switch m.mode {
	case modeRename:
		if msg.Type == tea.KeyEnter {
			err := m.store.Rename(getID(m.atCursor()), m.nameinput.Value())
			check(err)
			m.mode = modeNormal
		} else {
			var cmd tea.Cmd
			m.nameinput, cmd = m.nameinput.Update(msg)
			return cmd
		}
	case modeNormal:
		var pos task.Pos = task.Below
		switch msg.String() {
		case "alt+1":
			m.tabs.Set(0)
			m.updateTasks()
		case "alt+2":
			m.tabs.Set(1)
			m.updateTasks()
		case "j":
			m.setCursor(m.cursor + 1)
		case "k":
			m.setCursor(m.cursor - 1)
		case "i":
			m.edit()
		case "K":
			id := getID(m.atCursor())
			if m.moveSameParent(-1) {
				m.store.Move(getID(m.atCursor()), id, task.Above)
				m.updateTasks()
			}
		case "J":
			c := m.cursor
			id := getID(m.atCursor())
			if m.moveSameParent(1) {
				m.store.Move(id, getID(m.atCursor()), task.Above)
				m.updateTasks()
				m.setCursor(c)
				m.moveSameParent(1)
			}

		case "O":
			pos = task.Above
			fallthrough
		case "o":
			focused := getID(m.atCursor())
			id := task.RandomID()
			check(m.store.Create(id, m.now()))
			// we don't need to (and can't) move when there are no tasks yet
			if focused != "" {
				check(m.store.Move(id, focused, pos))
			}
			m.updateTasks()
			m.setCursor(m.cursor + int(pos))
			m.edit()
		}
	}
	return nil
}

func (m app) now() time.Time {
	return m.time.Add(time.Since(m.timeSetAt))
}

// updateTasks triggers a rerender of the viewport with all the tasks
func (m *app) updateTasks() {
	m.visible = traverse(m.store, "root")[1:]

	m.visible = m.filter(m.visible, m.predicates[m.tabs.Value()])
}

func (m *app) filter(paths []path, f predicate) []path {
	arr := []path{}
	// this makes sure that even if parent didn't pass the filter, it will still be displayed
	pile := []path{}
	for _, p := range paths {
		if len(pile) > 0 && len(p) <= len(pile[len(pile)-1]) {
			pile = []path{}
		}
		pile = append(pile, p)
		t := m.store.Get(getID(p))
		// still show those who have been created but don't meet the criteria
		if m.tabs.LastChanged().Before(t.Created) || f(t) {
			arr = append(arr, pile...)
			pile = []path{}
		}
	}
	return arr
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
	for i, path := range m.visible {
		var (
			bigspace bool
			title    lipgloss.Style = ui.TaskTitle
		)
		id := getID(path)
		t := m.store.Get(id)

		// style differences
		if m.store.GetParent(id) == "root" {
			bigspace = true
		} else {
			title = ui.SubTaskTitle
		}
		if i == m.cursor {
			title = title.Copy().Background(ui.Faded)
		}

		// renderer
		if bigspace {
			s += "\n"
		}
		s += strings.Repeat("   ", len(path)-2)
		s += ui.TaskIcon.Render(string(getIcon(t.Category)))
		switch {
		case m.mode == modeRename && m.cursor == i:
			s += m.nameinput.View()
		default:
			s += title.Render(t.Name)
		}
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
func (m app) atCursor() path {
	// if no items visible
	if m.cursor >= len(m.visible) {
		return []task.ID{}
	}
	return m.visible[m.cursor]
}

func getID(path path) task.ID {
	if len(path) < 1 {
		return ""
	}
	return path[len(path)-1]
}

func (m *app) edit() {
	m.mode = modeRename
	name := m.store.Get(getID(m.atCursor())).Name
	m.nameinput.SetValue(name)
	m.nameinput.SetCursor(len(name) - 1)
}

// this is required to move tasks around
// it needs to use the visible paths, otherwise it would be possible to
// move a task to a place that is NOT ON THE SCREEN
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

// this is required to move tasks around
// it needs to use the visible paths, otherwise it would be possible to
// move a task to a place that is NOT ON THE SCREEN
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

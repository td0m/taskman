package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/td0m/taskman/pkg/dateinput"
	"github.com/td0m/taskman/storage"
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

type predicate func(task.Task) bool

type app struct {
	viewport  viewport.Model
	dateinput dateinput.Model
	textinput textinput.Model

	tabs       ui.Tabs
	predicates []predicate

	mode mode

	all     task.Tasks
	storage *storage.JSONBackend

	visible []path
	cursor  int
}

// newApp creates a new taskman TUI app
func newApp() app {
	ti := textinput.NewModel()
	ti.Focus()
	ti.Prompt = ""
	ti.BackgroundColor = "#555"
	ti.TextColor = "#000"

	store := storage.NewJSON("./tasks.json")
	data, err := store.Fetch()
	if err != nil {
		panic(err)
	}

	today := time.Now().Truncate(time.Hour * 24)
	tomorrow := today.Add(time.Hour * 24)
	yday := today.Add(-time.Hour * 24)

	all := func(t task.Task) bool {
		return t.Done == nil || t.Done.After(today.Add(-time.Hour*24*5))
	}
	inbox := func(t task.Task) bool {
		return t.Due == nil
	}
	todayF := func(t task.Task) bool {
		return (t.Done == nil || t.Done.After(yday)) && (t.Due != nil && t.Due.Before(tomorrow))
	}

	return app{
		all:        data,
		storage:    store,
		viewport:   viewport.Model{},
		textinput:  ti,
		dateinput:  dateinput.NewModel(),
		tabs:       ui.NewTabs([]string{"All", "Inbox", "Today"}),
		predicates: []predicate{all, inbox, todayF},
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
		m.tabs.Width = msg.Width
		// on init:
		m.updateVisible()
		m.setCursor(m.cursor)
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			_, err := m.storage.Sync(m.all)
			if err != nil {
				fmt.Println(err)
			} else {
				return m, tea.Quit
			}
		}
		if msg.Type == tea.KeyEsc {
			m.mode = normalMode
		}
		if msg.Type == tea.KeyTab {
			c := m.cursor
			id := getID(m.atCursor())
			if m.moveSameParent(-1) {
				above := getID(m.atCursor())
				m.all.Move(id, above, "", task.Below)
				m.updateVisible()
				m.setCursor(c)
			}
		} else if msg.Type == tea.KeyShiftTab {
			c := m.cursor
			id := getID(m.atCursor())
			if m.moveUpLeft() {
				above := getID(m.atCursor())
				newParent := m.all.Parent[above]
				m.all.Move(id, newParent, above, 1)
				m.updateVisible()
				m.setCursor(c)
			}
		}
		switch m.mode {
		case titleMode:
			if msg.Type == tea.KeyEnter {
				m.mode = normalMode
				id := getID(m.atCursor())
				err := m.all.SetTitle(id, m.textinput.Value())
				if err != nil {
					panic(err)
				}
			} else {
				m.textinput, cmd = m.textinput.Update(msg)
				m.textinput.Width = len(m.textinput.Value()) + 1
				cmds = append(cmds, cmd)
			}
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
			anchor := task.Below
			switch msg.String() {
			case "alt+1":
				m.tabs.Set(0)
				m.setCursor(0)
				m.updateVisible()
			case "alt+2":
				m.tabs.Set(1)
				m.setCursor(0)
				m.updateVisible()
			case "alt+3":
				m.tabs.Set(2)
				m.setCursor(0)
				m.updateVisible()
			case tea.KeyEnter.String():
				id := getID(m.atCursor())
				t := m.all.Nodes[id]
				// TODO: handle error without panicking
				m.all.SetFolded(id, !t.Folded)
				// if err != nil {
				// 	panic(err)
				// }
				m.updateVisible()
			case "i":
				m.edit()
			case tea.KeyDelete.String():
				id := getID(m.atCursor())
				if len(id) > 0 {
					err := m.all.Remove(id)
					if err != nil {
						panic(err)
					}
					m.updateVisible()
					m.setCursor(m.cursor)
				}
			case "d":
				m.dateinput.SetValue(nil)
				m.mode = dateMode
			case "j", tea.KeyDown.String():
				m.setCursor(m.cursor + 1)
			case "k", tea.KeyUp.String():
				m.setCursor(m.cursor - 1)
			case "t":
				id := getID(m.atCursor())
				now := time.Now()
				var err error
				if m.all.Nodes[id].Done == nil {
					err = m.all.SetDone(id, &now)
				} else {
					err = m.all.SetDone(id, nil)
				}
				if err != nil {
					panic(err)
				}
				m.updateVisible()
			case "K":
				id := getID(m.atCursor())
				if m.moveSameParent(-1) {
					above := getID(m.atCursor())
					m.all.Move(above, m.all.Parent[id], id, task.Below)
					m.updateVisible()
				}
			case "J":
				c := m.cursor
				id := getID(m.atCursor())
				if m.moveSameParent(1) {
					above := getID(m.atCursor())
					m.all.Move(id, m.all.Parent[id], above, task.Below)
					m.updateVisible()
					m.setCursor(c)
					m.moveSameParent(1)
				}
			case "O":
				anchor = task.Above
				fallthrough
			case "o":
				id := getID(m.atCursor())
				parent := m.all.Parent[id]
				if len(parent) == 0 {
					parent = "root"
				}
				err := m.all.Add(parent, id, anchor)
				if err != nil {
					panic(err)
				}
				m.updateVisible()
				m.setCursor(m.cursor + int(anchor))
				m.edit()
			}
		}
	}
	m.viewport.SetContent(m.renderTasks())

	return m, tea.Batch(cmds...)
}

func (m *app) edit() {
	m.mode = titleMode
	t := m.all.Nodes[getID(m.atCursor())]
	m.textinput.SetValue(t.Title)
	m.textinput.Width = len(m.textinput.Value()) + 1
	m.textinput.SetCursor(m.textinput.Width)
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
	m.visible = traverse(m.all, "root")[1:]

	m.visible = m.filter(m.visible, m.predicates[m.tabs.Value()])
	// TODO: clamp cursor
	// m.setCursor(m.cursor) // for when we switch tabs and previous cursor is out of reach

	// save
	m.storage.Sync(m.all)

	sum, done := 0, 0
	for _, path := range m.visible {
		t := m.all.Nodes[getID(path)]
		// do not count those who are only there because its parent/child is
		if m.predicates[m.tabs.Value()](t) {
			if t.Done != nil {
				done++
			}
			sum++
		}
	}
	if sum > 0 {
		percent := done * 100 / sum
		m.tabs.Info = lipgloss.NewStyle().Foreground(ui.Secondary).Render(strconv.Itoa(percent) + "%")
	}
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
		t := m.all.Nodes[getID(p)]
		// still show those who have been created but don't meet the criteria
		if m.tabs.LastChanged().Before(t.Created) || f(t) {
			arr = append(arr, pile...)
			pile = []path{}
		}
	}
	return arr
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

// this function is needed to figure out the height of a task for scrolling to work properly
// this is because they can have varying heights due to custom spacing between groups
func (m app) sizeOf(i int) int {
	var (
		currentPath = m.visible[i]
		prevPath    path
		nextPath    path
	)
	if i > 0 {
		prevPath = m.visible[i-1]
	}
	if i+1 < len(m.visible) {
		nextPath = m.visible[i+1]
	}
	// rules for separating top level todos
	if len(currentPath) == 2 && ((len(currentPath) != len(nextPath) && len(nextPath) != 0) || len(currentPath) != len(prevPath)) {
		return 2
	}
	return 1
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
	return m.tabs.View() + m.viewport.View() + "\n" + statusline
}

func (m app) renderTasks() string {
	s := ""
	for i, currentPath := range m.visible {
		// s += strconv.Itoa(i) + "line\n"
		task := m.all.Nodes[getID(currentPath)]

		if m.sizeOf(i) == 2 {
			s += "\n"
		}

		// s +=
		if len(currentPath) > 2 {
			s += strings.Repeat("  ", len(currentPath)-2)
		}

		s += ui.RenderIcon(task)
		if m.mode == titleMode && i == m.cursor {
			s += m.textinput.View()
		} else {
			title := ui.Title(task)
			if i == m.cursor {
				if m.mode == normalMode {
					title = title.Copy().Background(ui.Faded).Foreground(ui.Background)
				}
			}
			if len(currentPath) == 2 {
				title = title.Copy().Bold(true)
			} else if len(currentPath) == 3 {
				title = title.Copy().Foreground(ui.Secondary)
			} else {
				title = title.Copy().Foreground(ui.Faded)
			}
			s += title.Render(task.Title)
		}
		if task.Done == nil {
			s += ui.RenderDue(task)
		}
		s += "\n"
	}
	return s
}

func traverse(m task.Tasks, id task.ID) []path {
	all := []path{{id}}
	if m.Nodes[id].Folded {
		return all
	}
	for _, child := range m.Children[id] {
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

package main

import (
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/td0m/taskman/internal/ui"
	"github.com/td0m/taskman/pkg/persist"
	"github.com/td0m/taskman/pkg/task"
	"github.com/td0m/taskman/pkg/task/date"
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

var (
	filePath = flag.String("file", "./tasks.json", "Path to task file")
)

func main() {
	flag.Parse()

	persist, err := persist.InJSON(*filePath)
	check(err)

	i := textinput.NewModel()
	i.Focus()
	i.Prompt = ""
	i.Width = 40

	a := &app{
		nameinput: i,
		viewport:  viewport.Model{},
		tabs:      ui.NewTabs([]string{"Outline", "Today", "Habits"}),
		time:      time.Now(),
		timeSetAt: time.Now(),

		store:   task.NewStore(),
		persist: persist,
	}

	a.predicates = []predicate{
		func(t task.Info) bool {
			if t.Repeats && a.now().Before(*t.NextDue()) {
				return false
			}
			return true
		},
		func(t task.Info) bool {
			due := t.NextDue()
			return due != nil && due.Before(date.StartOfDay(a.now().Add(time.Hour*24))) && !t.Done()
		},
		func(t task.Info) bool {
			return t.Repeats
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
	modeDue
	modeCategory
	modeClockIn
)

type predicate func(task.Info) bool

type app struct {
	mode   mode
	loaded bool

	viewport  viewport.Model
	nameinput textinput.Model
	tabs      ui.Tabs

	time         time.Time
	timeSetAt    time.Time
	tabChangedAt time.Time
	clockedInAt  time.Time

	cursor     int
	visible    []task.ID
	predicates []predicate

	store   task.StoreManager
	persist *persist.JSON
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

		if !m.loaded {
			var err error
			m.store, err = m.persist.Load()
			check(err)
			m.setTab(0) // needed to update tabLastChanged so that not all tasks are displayed
			m.updateTasks()
		}
		m.loaded = true
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
	var cmd tea.Cmd
	if m.mode == modeRename || m.mode == modeNormal {
		if msg.Type == tea.KeyTab {
			c := m.cursor
			id := m.atCursor()
			if m.moveSameParent(-1) {
				above := m.atCursor()
				m.store.Move(id, above, task.Into)
				m.updateTasks()
				m.setCursor(c)
			}
			return nil
		}
		if msg.Type == tea.KeyShiftTab {
			c := m.cursor
			id := m.atCursor()
			if m.moveUpLeft() {
				above := m.atCursor()
				m.store.Move(id, above, task.Below)
				m.updateTasks()
				m.setCursor(c)
			}
			return nil
		}
	}
	switch m.mode {
	case modeClockIn:
		if msg.Type == tea.KeyEnter || msg.String() == " " {
			m.clockOut()
		}
	case modeRename, modeDue, modeCategory:
		if msg.Type == tea.KeyEnter {
			switch m.mode {
			case modeRename:
				err := m.store.Rename(m.atCursor(), m.nameinput.Value())
				check(err)
			case modeDue:
				d, err := date.ParseDate(m.nameinput.Value())
				if err == nil {
					ds := []date.RepeatableDate{d}
					if m.nameinput.Value() == "" {
						ds = []date.RepeatableDate{}
					}
					check(m.store.SetDue(m.atCursor(), ds, m.now()))
				}
			case modeCategory:
				err := m.store.SetCategory(m.atCursor(), m.nameinput.Value())
				check(err)
			}
			m.mode = modeNormal
			m.save()
		} else {
			m.nameinput, cmd = m.nameinput.Update(msg)
		}
		m.nameinput.Width = len(m.nameinput.Value()) + 1
	case modeNormal:
		var pos task.Pos = task.Below
		switch msg.String() {
		case "g":
			m.setCursor(0)
		case "G":
			m.setCursor(len(m.visible))
		case "ctrl+d":
			m.setCursor(m.cursor + 10)
		case "ctrl+u":
			m.setCursor(m.cursor - 10)
		case "alt+1":
			m.setTab(0)
		case "alt+2":
			m.setTab(1)
		case "alt+3":
			m.setTab(2)
		case "j":
			m.setCursor(m.cursor + 1)
		case "k":
			m.setCursor(m.cursor - 1)
		case "i":
			if m.atCursor() != "" {
				m.edit()
			}
		case "d":
			if m.atCursor() != "" {
				m.editDue()
			}
		case " ":
			m.clockIn()
		case "t":
			if id := m.atCursor(); id != "" {
				check(m.store.Do(id, m.now()))
				m.save()
			}
		case tea.KeyDelete.String():
			if id := m.atCursor(); id != "" {
				check(m.store.Delete(id))
				m.updateTasks()
				m.setCursor(m.cursor) // make sure cursor is visible
			}
		case "r":
			if id := m.atCursor(); id != "" {
				t := m.store.Get(id)
				m.store.SetRepeats(id, !t.Repeats)
				m.save()
			}
		case "c":
			m.editCategory()
		case "K":
			id := m.atCursor()
			if m.moveSameParent(-1) {
				m.store.Move(m.atCursor(), id, task.Above)
				m.updateTasks()
			}
		case "J":
			c := m.cursor
			id := m.atCursor()
			if m.moveSameParent(1) {
				m.store.Move(id, m.atCursor(), task.Above)
				m.updateTasks()
				m.setCursor(c)
				m.moveSameParent(1)
			}
		case "O":
			pos = task.Above
			fallthrough
		case "o":
			focused := m.atCursor()
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
	return cmd
}

func (m app) now() time.Time {
	return m.time.Add(time.Since(m.timeSetAt))
}

func (m *app) setTab(i int) {
	m.tabs.Set(i)
	m.tabChangedAt = m.now()

	m.updateTasks()

	m.setCursor(0)
}

// updateTasks triggers a rerender of the viewport with all the tasks
func (m *app) updateTasks() {
	predicate := func(i task.Info) bool {
		return m.tabChangedAt.Before(i.Created) || m.predicates[m.tabs.Value()](i)
	}
	m.visible = filter(m.store, "root", predicate)
	m.save()
}

func (m *app) save() {
	check(m.persist.Save(m.store))
}

func filter(m task.StoreManager, id task.ID, f predicate) []task.ID {
	out := []task.ID{}
	for _, t := range m.GetChildren(id) {
		out = append(out, dfs(m, t, f)...)
	}
	return out
}

// dfs is a depth-first-search traversal utility
func dfs(store task.StoreManager, id task.ID, f predicate) []task.ID {
	out := []task.ID{}
	for _, child := range store.GetChildren(id) {
		children := dfs(store, child, f)
		out = append(out, children...)
	}
	if len(out) > 0 {
		return append([]task.ID{id}, out...)
	}
	if f(store.Get(id)) {
		return []task.ID{id}
	}
	return []task.ID{}
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
		current = m.visible[i]
	)
	// rules for separating top level todos
	if m.pathSize(current) == 2 {
		return 2
	}
	return 1
}

func (m app) pathSize(id task.ID) int {
	parent := m.store.GetParent(id)
	if parent == "root" {
		return 2
	}
	return 1 + m.pathSize(parent)
}

func (m app) viewTasks() string {
	s := ""
	for i, id := range m.visible {
		var (
			bigspace bool
			icon     rune           = '∙'
			title    lipgloss.Style = ui.TaskTitle
		)
		t := m.store.Get(id)
		pathSize := m.pathSize(id)

		// style differences
		if m.store.GetParent(id) == "root" {
			icon = getIcon(t.Category)
			bigspace = true
		} else {
			title = ui.SubTaskTitle
		}
		if i == m.cursor {
			title = title.Copy().Background(ui.Faded)
		}
		if t.Done() {
			title = title.Copy().Strikethrough(true)
		}

		// renderer
		if bigspace {
			s += "\n"
		}
		s += strings.Repeat("   ", pathSize-2)
		s += ui.TaskIcon.Render(string(icon))
		switch {
		case m.mode == modeRename && m.cursor == i:
			s += m.nameinput.View()
		default:
			s += title.Render(t.Name)
		}
		timeLogged := t.TimeLogged()
		if timeLogged > time.Duration(0) {
			s += formatDuration(timeLogged)
		}
		due := m.store.NextDue(id)
		if due != nil {
			s += ui.TaskDivider
			s += lipgloss.NewStyle().Foreground(m.getColor(*due)).Render(m.formatDate(*due))
			if t.Repeats {
				s += lipgloss.NewStyle().Padding(0, 1, 0, 1).Foreground(ui.Faded).Render("⭮")
			}
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
	statusline := ""
	switch m.mode {
	case modeDue:
		status := "✗"
		if d, err := date.ParseDate(m.nameinput.Value()); err == nil {
			status = "✓"
			if m.nameinput.Value() != "" {
				next := d.Next(m.now())
				status += " " + m.formatDate(next)
			}
		}
		statusline += m.nameinput.View() + status
	case modeCategory:
		status := "✗"
		if icon, ok := icons[m.nameinput.Value()]; ok {
			status = string(icon)
		}
		statusline += "category: " + m.nameinput.View() + status
	case modeClockIn:
		statusline += "clocked in: " + formatDuration(m.now().Sub(m.clockedInAt))
	}
	return m.tabs.View() + m.viewport.View() + "\n" + statusline
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

func (m app) atCursor() task.ID {
	// if no items visible
	if m.cursor >= len(m.visible) {
		return ""
	}
	return m.visible[m.cursor]
}

func (m *app) edit() {
	m.mode = modeRename
	name := m.store.Get(m.atCursor()).Name
	m.nameinput.SetValue(name)
	m.nameinput.Width = len(m.nameinput.Value()) + 1
	m.nameinput.SetCursor(len(name) - 1)
}

func (m *app) editDue() {
	m.mode = modeDue
	m.nameinput.SetValue("")
	m.nameinput.Width = len(m.nameinput.Value()) + 1
}

func (m *app) editCategory() {
	m.mode = modeCategory
	m.nameinput.SetValue(m.store.Get(m.atCursor()).Category)
	m.nameinput.Width = len(m.nameinput.Value()) + 1
}

// this is required to move tasks around
// it needs to use the visible paths, otherwise it would be possible to
// move a task to a place that is NOT ON THE SCREEN
func (m *app) moveSameParent(inc int) bool {
	if len(m.visible) == 0 {
		return false
	}
	path := m.pathSize(m.atCursor())
	all := m.visible
	i := m.cursor + inc
	for i >= 0 && i < len(all) {
		p := m.pathSize(all[i])
		// prevents from jumping to weird locations
		if p < path {
			return false
		}
		if p == path {
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
	path := m.pathSize(m.visible[m.cursor])
	before := m.visible[:m.cursor]
	for i := len(before) - 1; i >= 0; i-- {
		p := m.pathSize(before[i])
		if p < path {
			m.setCursor(i)
			return true
		}
	}
	return false
}
func (m app) getColor(t time.Time) lipgloss.Color {
	diff := t.Sub(date.StartOfDay(m.now()))
	switch days := int(diff.Hours()) / 24; {
	case days <= 2:
		return ui.Red
	case days <= 14:
		return ui.Orange
	default:
		return ui.Faded
	}
}

func (m app) formatDate(t time.Time) string {
	now := m.now().Truncate(time.Hour * 24)
	diff := t.Sub(now)
	switch days := int(diff.Hours()) / 24; {
	case days == 0:
		return "today"
	case days < 14:
		suffix := ""
		if days > 1 {
			suffix = "s"
		}
		return strconv.Itoa(days) + " day" + suffix
	// max 1 month
	case days <= 31:
		return strconv.Itoa(days/7) + " weeks"
	// months
	default:
		postfix := ""
		months := days / 31
		if months > 1 {
			postfix = "s"
		}
		return strconv.Itoa(months) + " month" + postfix
	}
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	style := ui.TaskTimer
	bracket := lipgloss.NewStyle().Foreground(ui.Faded).Render
	return bracket(" (") + style.Render(d.String()) + bracket(") ")
}

func (m *app) clockIn() {
	m.mode = modeClockIn
	m.clockedInAt = m.now()
}

func (m *app) clockOut() {
	m.mode = modeNormal
	check(m.store.Log(m.atCursor(), m.clockedInAt, m.now()))
	m.save()
}

package dateinput

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	indicator = lipgloss.NewStyle().Padding(0, 1).Bold(true)
	checkmark = indicator.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: "#00ad3b", Dark: "#73F59F"}).
			Render("✓")

	cross = indicator.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "", Dark: "#FF5047"}).
		Render("✗")

	faded = lipgloss.AdaptiveColor{Light: "#666", Dark: "#999"}
)

type Model struct {
	i     textinput.Model
	value *time.Time
}

func NewModel() Model {
	i := textinput.NewModel()
	i.Focus()
	i.CharLimit = 20
	i.Prompt = ""
	return Model{
		i: i,
	}
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.i, cmd = m.i.Update(msg)
		m.value = parseDate(m.i.Value())
		return m, cmd
	}
	return m, nil
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m *Model) View() string {
	indicator := cross
	if m.i.Value() == "" {
		indicator = ""
	} else if m.value != nil {
		indicator = checkmark + " " + format(*m.value)
	}
	prefix := "due"
	return lipgloss.NewStyle().Foreground(faded).Render(prefix+": ") + m.i.View() + "" + indicator
}

func (m *Model) Value() *time.Time {
	return m.value
}
func (m *Model) SetValue(t *time.Time) {
	m.value = t
	if t == nil {
		m.i.SetValue("")
		return
	}
	m.i.SetValue((*t).Format(formats[0]))
}

func parseDate(s string) *time.Time {
	// s = strings.ToLower(s)
	// if s == "" {
	// 	return nil
	// }
	// for _, fmt := range formats {
	// 	t, err := time.Parse(fmt, s)
	// 	if err == nil {
	// 		return &t
	// 	}
	// }
	today := time.Now().Truncate(time.Hour * 24)
	{
		for i, fmt := range []string{"today", "tomorrow"} {
			end := min(len(s), len(fmt))
			if s == fmt[:end] {
				tom := today.Add(time.Hour * 24 * time.Duration(i))
				return &tom
			}
		}
	}
	for i := time.Sunday; i <= time.Saturday; i++ {
		fmt := strings.ToLower(i.String())
		end := min(len(s), len(fmt)-1)
		if s == fmt[:end] {
			d := nextWeekday(today, i)
			return &d
		}
	}
	duration, err := parseRelative(s)
	if err == nil {
		d := today.Add(duration)
		return &d
	}
	r := regexp.MustCompile(`([0-9])(st|nd|rd|th)`)
	s = string(r.ReplaceAll([]byte(s), []byte("$1")))
	t, err := parseAbsolute(s, today)
	if err == nil {
		return &t
	}
	return nil
}

func nextWeekday(t time.Time, d time.Weekday) time.Time {
	day := d - t.Weekday()
	if day < 0 {
		day += 7
	}
	return t.Add(time.Duration(day) * time.Hour * 24)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func format(t time.Time) string {
	now := time.Now().Truncate(time.Hour * 24)
	diff := t.Sub(now)
	switch days := int(diff.Hours()) / 24; {
	case days < 14:
		return strconv.Itoa(days) + " days"
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

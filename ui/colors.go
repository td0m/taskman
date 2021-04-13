package ui

import (
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/td0m/taskman/task"
)

const (
	Background = lipgloss.Color("#000")

	Primary   = lipgloss.Color("#fff")
	Secondary = lipgloss.Color("#888")
	Faded     = lipgloss.Color("#555")

	Green  = lipgloss.Color("#00a352")
	Red    = lipgloss.Color("#c42912")
	Yellow = lipgloss.Color("#c4b810")
	Orange = lipgloss.Color("#c27510")
)

var (
	icon   = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	undone = icon.Copy().Foreground(Secondary).Render("•")
	done   = icon.Copy().Foreground(Green).Render("✓")

	title     = lipgloss.NewStyle().Bold(true)
	titleDone = title.Copy().Foreground(Secondary).Strikethrough(true)

	due       = lipgloss.NewStyle().Foreground(Secondary)
	dueSoon   = due.Copy().Foreground(Red)
	dueYellow = due.Copy().Foreground(Yellow)
	dueOrange = due.Copy().Foreground(Orange)

	divider = lipgloss.NewStyle().Padding(0, 1).Foreground(Faded).Render("•")
)

func RenderIcon(t task.Task) string {
	icon := undone
	if t.Done != nil {
		icon = done
	}
	return icon
}

func Title(t task.Task) lipgloss.Style {
	if t.Done != nil {
		return titleDone
	}

	return title
}

func RenderDue(t task.Task) string {
	if t.Due == nil {
		return ""
	}
	f := due
	days := (*t.Due).Sub(time.Now().Truncate(time.Hour*24)).Hours() / 24
	if days < 1 {
		f = dueSoon
	} else if days == 1 {
		f = dueOrange
	} else if days < 14 {
		f = dueYellow
	}
	return divider + f.Render(dateToString(*t.Due))
}

func dateToString(t time.Time) string {
	now := time.Now().Truncate(time.Hour * 24)
	diff := t.Sub(now)
	switch days := int(diff.Hours()) / 24; {
	case days < 0:
		return "overdue"
	case days == 0:
		return "today"
	case days == 1:
		return "1 day"
	case days < 14:
		return strconv.Itoa(days) + " days"
	// max 1 month
	case days < 31:
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

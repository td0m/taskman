package task

import (
	"time"

	"github.com/td0m/taskman/pkg/task/date"
)

type ID string

type Info struct {
	// constants
	ID ID

	// behavioural properties
	Name        string
	Repeats     bool
	Due         []date.RepeatableDate
	DoneHistory []time.Time
	ClockIns    []ClockIn
	Category    string
	Notes       string

	// visibility properties
	Archived bool
	Folded   bool
}

type Task struct {
	Info
	Parent   *Task
	Children []*Task
}

func (t Task) Done() bool {
	panic("unimplemented")
}

type ClockIn struct {
	Start time.Time
	End   time.Time
}

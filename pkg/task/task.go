package task

import (
	"math/rand"
	"time"

	"github.com/td0m/taskman/pkg/task/date"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type ID string

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(l int) string {
	b := make([]byte, l)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
func RandomID() ID {
	return ID(randomString(8))
}

type Info struct {
	// constants
	Created time.Time

	// behavioural properties
	Name        string
	Repeats     bool
	Due         []date.RepeatableDate
	DueChanged  *time.Time
	DoneHistory []time.Time
	ClockIns    []ClockIn
	Category    string
	Notes       string

	// visibility properties
	Archived bool
	Folded   bool
}

func min(a, b time.Time) time.Time {
	if b.Before(a) {
		return b
	}
	return a
}

const unixToInternal int64 = (1969*365 + 1969/4 - 1969/100 + 1969/400) * 24 * 60 * 60

// NextDue computes the next due date for a task
// it gets all next due dates and returns the smallest one
func (t Info) NextDue() *time.Time {
	if len(t.Due) == 0 || t.DueChanged == nil {
		return nil
	}
	// max time that still works with comparisons
	earliest := time.Unix(1<<63-1-unixToInternal, 999999999)
	for _, d := range t.Due {
		due := d.Next(*t.DueChanged)
		earliest = min(earliest, due)
	}
	return &earliest
}

func (t Info) LastDone() *time.Time {
	all := t.DoneHistory
	if len(all) == 0 {
		return nil
	}
	return &all[len(all)-1]
}

func (t Info) Done() bool {
	return !(t.Repeats || t.LastDone() == nil)
}

type Task struct {
	Info
	Parent   *Task
	Children []*Task
}

type ClockIn struct {
	Start time.Time
	End   time.Time
}

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

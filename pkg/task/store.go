package task

import (
	"time"

	"github.com/td0m/taskman/pkg/task/date"
)

type Pos int

const (
	Above Pos = iota
	Below
)

type Store interface {
	Rename(ID, string) error
	SetCategory(ID, string) error
	Done(ID, time.Time) error
	UndoLastDone(ID) error
	SetDue(ID, []date.RepeatableDate) error
	SetRepeats(ID, bool) error

	Clock(ID, time.Time, time.Time) error

	Move(target, parent, anchor ID, pos Pos) error
	Archive(ID) error

	All() []Task
}

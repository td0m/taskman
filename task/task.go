package task

import "time"

type ID string

type Tasks map[ID]Task

type Task struct {
	Title   string
	Created time.Time
	Done    *time.Time
	Due     *time.Time

	Folded bool

	Children []ID
}

func newTask() Task {
	return Task{
		Created: time.Now(),
	}
}

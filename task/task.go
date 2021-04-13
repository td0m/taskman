package task

import "time"

type ID string

type Tasks map[ID]Task

type Task struct {
	Title string
	Done  *time.Time
	Due   *time.Time

	Folded bool

	Children []ID
}

func newTask() Task {
	return Task{}
}

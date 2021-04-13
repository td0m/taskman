package task

import "time"

type ID string

type Tasks map[ID]Task

type Task struct {
	Title string
	Done  *time.Time
	Due   *time.Time

	Children []ID
}

func newTask() Task {
	return Task{}
}

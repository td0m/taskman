package task

import "time"

type ID string

type Tasks struct {
	Nodes    map[ID]Task `json:"nodes"`
	Children map[ID][]ID `json:"children"`
	Parent   map[ID]ID   `json:"parent"`
}

func NewTasks() Tasks {
	return Tasks{
		Nodes:    map[ID]Task{"root": {}},
		Children: map[ID][]ID{},
		Parent:   map[ID]ID{},
	}
}

type Task struct {
	Title   string     `json:"title,omitempty"`
	Created time.Time  `json:"created,omitempty"`
	Done    *time.Time `json:"done,omitempty"`
	Due     *time.Time `json:"due,omitempty"`

	Folded bool `json:"folded,omitempty"`
}

func newTask() Task {
	return Task{
		Created: time.Now(),
	}
}

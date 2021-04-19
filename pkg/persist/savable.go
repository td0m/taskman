package persist

import (
	"errors"

	"github.com/td0m/taskman/pkg/task"
)

type ID task.ID

// the default data structure for tasks uses parent and children pointers
// we cannot store pointers in json, but we still need a method to quickly save/load them
type savable struct {
	Nodes    map[ID]task.Info
	Parent   map[ID]ID
	Children map[ID][]ID
}

// newSavable parses a list of tasks (that use pointers) and transforms it into a serializable data structure
func newSavable(ts []task.Task) (savable, error) {
	s := savable{
		Nodes:    map[ID]task.Info{},
		Parent:   map[ID]ID{},
		Children: map[ID][]ID{},
	}

	for _, t := range ts {
		s.Nodes[ID(t.ID)] = t.Info
		parent := ID("")
		if t.Parent != nil {
			parent = ID(t.Parent.ID)
		}
		s.Parent[ID(t.ID)] = parent
		children := make([]ID, len(t.Children))
		for i, c := range t.Children {
			children[i] = ID(c.ID)
		}
		s.Children[ID(t.ID)] = children
	}

	return s, s.check()
}

// Load loads a serializable savable struct into a list of tasks that use pointers
// it also validates it
func (s savable) Load() ([]task.Task, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	tasks := make([]task.Task, len(s.Nodes))
	indexes := map[ID]int{}
	i := 0
	for id, t := range s.Nodes {
		tasks[i] = task.Task{
			Info: t,
		}
		indexes[id] = i
		i++
	}
	for i := range tasks {
		t := tasks[i]
		parentID := s.Parent[ID(t.ID)]
		if len(parentID) != 0 {
			t.Parent = &tasks[indexes[parentID]]
		}
		childrenIDs := s.Children[ID(t.ID)]
		children := make([]*task.Task, len(childrenIDs))
		for i, id := range childrenIDs {
			children[i] = &tasks[indexes[id]]
		}
		t.Children = children
		tasks[i] = t
	}

	return tasks, nil
}

// check checks whether the savable data structure is valid
// TODO: move this to the store itself, since that's probably where the core logic should be
func (s savable) check() error {
	rootFound := false
	for _, p := range s.Parent {
		if len(p) == 0 {
			if rootFound {
				return errors.New("more than 1 root")
			}
			rootFound = true
		}
	}
	return nil
}

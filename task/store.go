package task

import (
	"errors"
	"math/rand"
	"time"
)

var (
	ErrNoParent = errors.New("invalid parent ID")
	ErrBadID    = errors.New("invalid ID")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandID() ID {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return ID(b)
}

func (tasks Tasks) Add(parent ID, cursor ID, anchor int) error {
	p, found := tasks[parent]
	if !found {
		return ErrNoParent
	}
	newID := RandID()
	t := newTask()
	tasks[newID] = t
	for i, id := range p.Children {
		if id == cursor {
			p.Children = insert(p.Children, i+anchor, newID)
			tasks[parent] = p
			return nil
		}
	}
	p.Children = append(p.Children, newID)
	tasks[parent] = p
	return nil
}

func (tasks Tasks) SetTitle(id ID, title string) error {
	t, found := tasks[id]
	if !found {
		return ErrBadID
	}
	t.Title = title
	tasks[id] = t
	return nil
}

func (tasks Tasks) SetDone(id ID, done *time.Time) error {
	t, found := tasks[id]
	if !found {
		return ErrBadID
	}
	t.Done = done
	tasks[id] = t
	return nil
}

func (tasks Tasks) SetDue(id ID, due *time.Time) error {
	t, found := tasks[id]
	if !found {
		return ErrBadID
	}
	t.Due = due
	tasks[id] = t
	return nil
}

func (tasks Tasks) SetFolded(id ID, folded bool) error {
	t, found := tasks[id]
	if !found {
		return ErrBadID
	}
	if len(t.Children) == 0 {
		return errors.New("cannot fold empty")
	}
	t.Folded = folded
	tasks[id] = t
	return nil
}

func (tasks Tasks) Remove(id ID, parent ID) error {
	p, ok := tasks[parent]
	if !ok {
		return errors.New("parent not found")
	}
	for i := range p.Children {
		if p.Children[i] == id {
			p.Children = remove(p.Children, i)
			tasks[parent] = p
			return nil
		}
	}
	return errors.New("task not found")
}

func remove(s []ID, i int) []ID {
	return append(s[:i], s[i+1:]...)
}

func (tasks Tasks) Move(id ID, oldParent ID, parent ID, cursor ID, anchor int) error {
	p, ok := tasks[parent]
	if !ok {
		return errors.New("parent not found")
	}
	remove := func() error {
		return tasks.Remove(id, oldParent)
	}
	for i, idd := range p.Children {
		if idd == cursor {
			p.Children = insert(p.Children, i+anchor, id)
			tasks[parent] = p
			return remove()
		}
	}
	// if cursor not found
	p.Children = append(p.Children, id)
	tasks[parent] = p
	return remove()

}

func insert(a []ID, index int, value ID) []ID {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

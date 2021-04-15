package task

import (
	"errors"
	"math/rand"
	"time"
)

type Pos int

const (
	Above Pos = iota
	Below
)

var (
	ErrNoParent = errors.New("invalid parent ID")
	ErrBadID    = errors.New("invalid ID")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomID() ID {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return ID(b)
}

func (t *Tasks) Add(parent ID, anchor ID, pos Pos) error {
	id := randomID()
	t.Nodes[id] = newTask()
	t.Move(id, parent, anchor, pos)
	return nil
}

func (t *Tasks) Move(target ID, parent ID, anchor ID, pos Pos) {
	// delete from current parent
	if parent, ok := t.Parent[target]; ok {
		t.removeChild(parent, target)
	}
	// update current parent
	t.Parent[target] = parent

	children := t.Children[parent]
	for i, c := range children {
		if c == anchor {
			t.Children[parent] = insert(children, i+int(pos), target)
			return
		}
	}
	// cursor not found but still insert
	t.Children[parent] = append(children, target)
}

func (t *Tasks) removeChild(parent, child ID) {
	// remove parent
	delete(t.Parent, child)
	// remove child
	children := t.Children[parent]
	for i, c := range children {
		if c == child {
			t.Children[parent] = append(children[:i], children[i+1:]...)
			return
		}
	}
}

func (tasks Tasks) SetTitle(id ID, title string) error {
	t, found := tasks.Nodes[id]
	if !found {
		return ErrBadID
	}
	t.Title = title
	tasks.Nodes[id] = t
	return nil
}

func (tasks *Tasks) SetDone(id ID, done *time.Time) error {
	t, found := tasks.Nodes[id]
	if !found {
		return ErrBadID
	}
	t.Done = done

	for _, c := range tasks.Children[id] {
		err := tasks.SetDone(c, done)
		if err != nil {
			return err
		}
	}
	tasks.Nodes[id] = t
	return nil
}

func (tasks Tasks) SetDue(id ID, due *time.Time) error {
	t, found := tasks.Nodes[id]
	if !found {
		return ErrBadID
	}
	t.Due = due
	tasks.Nodes[id] = t
	return nil
}

func (tasks Tasks) SetFolded(id ID, folded bool) error {
	t, found := tasks.Nodes[id]
	if !found {
		return ErrBadID
	}
	// if len(t.Children) == 0 {
	// 	return errors.New("cannot fold empty")
	// }
	t.Folded = folded
	tasks.Nodes[id] = t
	return nil
}

func (tasks *Tasks) Remove(id ID) error {
	// delete itself
	delete(tasks.Nodes, id)

	// delete all children recursively
	for _, c := range tasks.Children[id] {
		tasks.Remove(c)
	}
	delete(tasks.Children, id)

	// delete from parent
	tasks.removeChild(tasks.Parent[id], id)
	return nil
}

func insert(a []ID, index int, value ID) []ID {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

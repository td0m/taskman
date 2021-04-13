package task

import (
	"errors"
	"math/rand"
	"time"
)

var (
	ErrNoParent = errors.New("invalid parent ID")
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
	id := RandID()
	tasks[id] = newTask()
	for i, id := range p.Children {
		if id == cursor {
			p.Children = insert(p.Children, i+anchor, id)
			tasks[parent] = p
			return nil
		}
	}
	p.Children = append(p.Children, id)
	tasks[parent] = p
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

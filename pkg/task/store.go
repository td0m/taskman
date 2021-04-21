package task

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/td0m/taskman/pkg/task/date"
)

type Pos int

const (
	Above Pos = iota
	Below
	Into
)

type Serializable interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}

type StoreManager interface {
	Serializable

	Create(ID, time.Time) error
	Rename(ID, string) error
	SetCategory(ID, string) error
	Do(ID, time.Time) error
	UnDo(ID) error
	SetDue(ID, []date.RepeatableDate, time.Time) error
	SetRepeats(ID, bool) error

	Clock(ID, time.Time, time.Time) error

	Move(target, anchor ID, pos Pos) error
	Archive(ID) error

	Root() *Task
	Get(ID) Info
	GetChildren(ID) []ID
	GetParent(ID) ID

	NextDue(ID) *time.Time
}

var _ StoreManager = &Store{}

var (
	ErrIDAlreadyExists = errors.New("task with the given ID already exists")
	ErrNotFound        = errors.New("not found")
	ErrParentNotFound  = errors.New("parent not found")
	ErrAnchorNotFound  = errors.New("anchor not found")
)

// the default data structure for tasks uses parent and children pointers
// we cannot store pointers in json, but we still need a method to quickly save/load them
type Store struct {
	Nodes    map[ID]Info
	Parent   map[ID]ID
	Children map[ID][]ID
}

func NewStore() *Store {
	return &Store{
		Nodes:    map[ID]Info{"root": {Created: time.Now()}},
		Parent:   map[ID]ID{},
		Children: map[ID][]ID{},
	}
}

// Check validates a given json file
func (s Store) Check() error {
	return s.check()
}

func (s *Store) MarshalJSON() ([]byte, error) {
	type alias Store
	return json.Marshal(alias(*s))
}

func (s *Store) UnmarshalJSON(bs []byte) error {
	type alias Store
	var out alias
	err := json.Unmarshal(bs, &out)
	s.Nodes = out.Nodes
	s.Children = out.Children
	s.Parent = out.Parent
	return err
}

// check checks whether the savable data structure is valid
// TODO: move this to the store itself, since that's probably where the core logic should be
func (s Store) check() error {
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

func (s *Store) addChild(parent, child ID) error {
	children := s.Children[parent]
	s.Children[parent] = append(children, child)
	return nil
}

func (s *Store) Create(id ID, created time.Time) error {
	_, found := s.Nodes[id]
	if found {
		return ErrIDAlreadyExists
	}
	s.Nodes[id] = Info{
		Created: created,
	}
	s.Parent[id] = "root"
	return s.addChild("root", id)
}

func (s *Store) Rename(id ID, name string) error {
	t, ok := s.Nodes[id]
	if !ok {
		return ErrNotFound
	}
	t.Name = name
	s.Nodes[id] = t
	return nil
}

func (s *Store) SetCategory(_ ID, _ string) error {
	panic("not implemented") // TODO: Implement
}

func (s *Store) recalculateDue(id ID, t time.Time) error {
	return nil
}

func (s *Store) Do(_ ID, _ time.Time) error {
	panic("not implemented") // TODO: Implement
}

func (s *Store) UnDo(_ ID) error {
	panic("not implemented") // TODO: Implement
}

// SetDue sets the due date of a node
// returns an error if task not found
func (s *Store) SetDue(id ID, due []date.RepeatableDate, time time.Time) error {
	node, ok := s.Nodes[id]
	if !ok {
		return ErrNotFound
	}
	node.Due = due
	node.DueChanged = &time
	s.Nodes[id] = node
	return s.recalculateDue(id, time)
}

func (s *Store) SetRepeats(_ ID, _ bool) error {
	panic("not implemented") // TODO: Implement
}

func (s *Store) Clock(_ ID, _ time.Time, _ time.Time) error {
	panic("not implemented") // TODO: Implement
}

func (s *Store) detach(child ID) error {
	parent, ok := s.Parent[child]
	if !ok {
		return ErrParentNotFound
	}
	// remove parent
	delete(s.Parent, child)
	// remove child
	children := s.Children[parent]
	for i, c := range children {
		if c == child {
			s.Children[parent] = append(children[:i], children[i+1:]...)
			return nil
		}
	}
	return errors.New("child not found! there is a bug somewhere in your code that updates the parent but not child")
}

func (s *Store) attach(parent, child ID, i int) error {
	if err := s.detach(child); err != nil {
		return err
	}
	s.Parent[child] = parent
	s.Children[parent] = insert(s.Children[parent], i, child)
	return nil
}

func (s *Store) Move(target ID, anchor ID, pos Pos) error {
	{
		_, ok := s.Nodes[target]
		if !ok {
			return ErrNotFound
		}
		_, ok = s.Nodes[anchor]
		if !ok {
			return ErrAnchorNotFound
		}
	}
	switch pos {
	case Into:
		s.attach(anchor, target, len(s.Children[anchor]))
	case Below, Above:
		anchorParent, ok := s.Parent[anchor]
		if !ok {
			return ErrParentNotFound
		}
		// update current parent
		children := s.Children[anchorParent]
		for i, c := range children {
			if c == anchor {
				return s.attach(anchorParent, target, i+int(pos))
			}
		}
		return errors.New("internal bug! this should never happen, since parent should always contain its children")
	}
	return nil
}

func (s *Store) Archive(_ ID) error {
	panic("not implemented") // TODO: Implement
}

func (s *Store) get(id ID) *Task {
	task := &Task{Info: s.Nodes[id]}
	for _, c := range s.Children[id] {
		child := s.get(c)
		child.Parent = task
		task.Children = append(task.Children, child)
	}
	return task
}

func (s *Store) Root() *Task {
	return s.get("root")
}

func insert(a []ID, index int, value ID) []ID {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

func contains(ids []ID, search ID) bool {
	for _, id := range ids {
		if id == search {
			return true
		}
	}
	return false
}

func (s *Store) Get(id ID) Info {
	return s.Nodes[id]
}

func (s *Store) GetChildren(id ID) []ID {
	return s.Children[id]
}
func (s *Store) GetParent(id ID) ID {
	return s.Parent[id]
}

func (s *Store) NextDue(id ID) *time.Time {
	node, ok := s.Nodes[id]
	if !ok {
		return nil
	}
	due := node.NextDue()
	if parentDue := s.NextDue(s.Parent[id]); parentDue != nil && (due == nil || (*parentDue).Before(*due)) {
		return parentDue
	}
	return due
}

package task

import (
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/td0m/taskman/pkg/task/date"
)

func TestStore_Create(t *testing.T) {
	is := is.New(t)

	s := NewStore()
	// root must already exist
	err := s.Create("root", time.Now())
	is.Equal(err, ErrIDAlreadyExists)
	is.Equal(len(dfs(s.Root())), 1)

	// new task without error
	now := time.Now()
	is.NoErr(s.Create("goal", now))
	is.Equal(len(dfs(s.Root())), 2)
	is.Equal(s.Root().Children[0].Parent, s.Root())
	is.Equal(s.Root().Children[0].Created, now)
	// creating another task with the same id fails
	is.Equal(s.Create("goal", time.Now()), ErrIDAlreadyExists)
}

func TestStore_Rename(t *testing.T) {
	s := NewStore()
	s.Create("goal", time.Now())

	t.Run("renames a valid task", func(t *testing.T) {
		is := is.New(t)
		is.NoErr(s.Rename("goal", "hello"))
		is.Equal(s.Root().Children[0].Name, "hello")
	})

	t.Run("returns error on invalid task ID", func(t *testing.T) {
		is := is.New(t)
		is.Equal(s.Rename("invalid", "hello"), ErrNotFound)
	})
}

func TestStore_Move(t *testing.T) {
	is := is.New(t)
	var s StoreManager = NewStore()
	s.Create("goals", time.Now())
	s.Create("deadline", time.Now())

	t.Run("cannot move root", func(t *testing.T) {
		is := is.New(t)
		err := s.Move("root", "deadline", Below)
		is.Equal(err, ErrParentNotFound)
	})

	t.Run("fails on invalid child", func(t *testing.T) {
		is := is.New(t)
		err := s.Move("invalid", "deadline", Below)
		is.Equal(err, ErrNotFound)
	})
	t.Run("fails on invalid anchor", func(t *testing.T) {
		is := is.New(t)
		err := s.Move("goals", "invalid", Below)
		is.Equal(err, ErrAnchorNotFound)
	})

	t.Run("respects position", func(t *testing.T) {
		t.Run("into", func(t *testing.T) {
			is.NoErr(s.Create("deadline1", time.Now()))
			is.NoErr(s.Create("deadline2", time.Now()))
			is.NoErr(s.Create("deadline3", time.Now()))
			is.NoErr(s.Move("deadline1", "deadline", Into))
			is.NoErr(s.Move("deadline2", "deadline", Into))
			is.NoErr(s.Move("deadline3", "deadline", Into))
			is.Equal(s.GetChildren("deadline"), []ID{"deadline1", "deadline2", "deadline3"})

		})
		t.Run("above", func(t *testing.T) {
			is := is.New(t)
			is.NoErr(s.Create("deadline2.5", time.Now()))
			err := s.Move("deadline2.5", "deadline3", Above)
			is.NoErr(err)
			is.Equal(s.GetChildren("deadline"), []ID{"deadline1", "deadline2", "deadline2.5", "deadline3"})
		})
		t.Run("below", func(t *testing.T) {
			is := is.New(t)
			is.NoErr(s.Create("deadline1.5", time.Now()))
			err := s.Move("deadline1.5", "deadline1", Below)
			is.NoErr(err)
			is.Equal(s.GetChildren("deadline"), []ID{"deadline1", "deadline1.5", "deadline2", "deadline2.5", "deadline3"})
		})
	})
	t.Run("cannot put above/below root", func(t *testing.T) {
		is := is.New(t)
		is.NoErr(s.Create("root2", time.Now()))
		err := s.Move("root2", "root", Below)
		is.True(err != nil)
	})
}

func TestStore_SetDue(t *testing.T) {
	var s StoreManager = NewStore()

	start := time.Time{}                 // monday
	wed := start.Add(time.Hour * 24 * 2) // wednesday
	t.Run("calcualtes NextDate", func(t *testing.T) {
		is := is.New(t)

		s.Create("deadline", start)
		err := s.SetDue("deadline", []date.RepeatableDate{date.NewWeekday(time.Wednesday)}, start)
		is.NoErr(err)
		is.Equal(*s.Get("deadline").NextDue(), wed)

		// err = s.Do("deadline", start)
		// is.NoErr(err)
	})

	t.Run("(recursive) parent's due date is the maximum due date for all children", func(t *testing.T) {
		is := is.New(t)
		s.Create("deadline1", start)
		s.Create("deadline1.1", start)
		is.NoErr(s.Move("deadline1", "deadline", Into))
		is.NoErr(s.Move("deadline1.1", "deadline1", Into))
		s.SetDue("deadline1.1", []date.RepeatableDate{date.NewWeekday(time.Thursday)}, start)
		is.Equal(*s.NextDue("deadline1.1"), wed)
	})
	t.Run("no due inside of an item with due will return parent's due date", func(t *testing.T) {
		is := is.New(t)
		is.Equal(*s.NextDue("deadline1"), wed)
	})
}

// dfs is a depth-first-search traversal utility
// it is used to compare trees
func dfs(t *Task) []Task {
	out := []Task{*t}
	for _, child := range t.Children {
		out = append(out, dfs(child)...)
	}
	return out
}

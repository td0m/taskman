package task

import (
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/td0m/taskman/pkg/task/date"
)

// dfs is a depth-first-search traversal utility
// it is used to compare trees
func dfs(t *Task) []Task {
	out := []Task{*t}
	for _, child := range t.Children {
		out = append(out, dfs(child)...)
	}
	return out
}

func setup() StoreManager {
	var s StoreManager = NewStore()
	time := time.Time{}
	s.Create("foo", time)
	s.Create("foo1", time)
	s.Create("foo1.1", time)
	s.Create("foo1.1.1", time)
	s.Create("foo2", time)

	s.Create("daily routine", time)
	s.SetDue("daily routine", []date.RepeatableDate{date.NewDayOffset(1)}, time)
	s.SetRepeats("daily routine", true)

	s.Create("weekly routine", time)
	s.SetDue("weekly routine", []date.RepeatableDate{date.NewDayOffset(7)}, time)
	s.SetRepeats("weekly routine", true)

	s.Move("foo1", "foo", Into)
	s.Move("foo1.1", "foo1", Into)
	s.Move("foo1.1.1", "foo1.1", Into)
	s.Move("foo2", "foo", Into)
	return s
}

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
	t.Run("moving an undone task to a done parent undoes the parent", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		is.NoErr(s.Do("foo1", time.Time{}))
		is.True(s.Get("foo1").Done())
		is.NoErr(s.Move("foo2", "foo1", Into))
		is.True(!s.Get("foo1").Done())
	})
	// TODO: test that moving applies rules for: category, repeats
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

func TestStore_SetCategory(t *testing.T) {
	is := is.New(t)
	var s StoreManager = NewStore()
	time := time.Time{}
	s.Create("foo", time)
	s.Create("foo1", time)
	s.Create("foo1.1", time)
	s.Create("foo1.1.1", time)
	s.Create("foo2", time)

	s.Move("foo1", "foo", Into)
	s.Move("foo1.1", "foo1", Into)
	s.Move("foo1.1.1", "foo1.1", Into)
	s.Move("foo2", "foo", Into)

	t.Run("throws error if id not found", func(t *testing.T) {
		is := is.New(t)
		err := s.SetCategory("bar", "category")
		is.Equal(err, ErrNotFound)
	})
	t.Run("fails to set category of root", func(t *testing.T) {
		is := is.New(t)
		err := s.SetCategory("root", "category")
		is.True(err != nil)
	})
	t.Run("sets category of item and all of its children", func(t *testing.T) {
		is := is.New(t)
		err := s.SetCategory("foo", "bar")
		fooAndChildren := dfs(s.Root().Children[0])
		for _, t := range fooAndChildren {
			is.Equal(t.Category, "bar")
		}
		is.NoErr(err)
	})
}

func TestStore_SetRepeats(t *testing.T) {
	is := is.New(t)
	var s StoreManager = NewStore()
	time := time.Time{}
	s.Create("foo", time)
	s.Create("foo1", time)
	s.Create("foo1.1", time)
	s.Create("foo1.1.1", time)
	s.Create("foo2", time)

	s.Move("foo1", "foo", Into)
	s.Move("foo1.1", "foo1", Into)
	s.Move("foo1.1.1", "foo1.1", Into)
	s.Move("foo2", "foo", Into)

	t.Run("throws error if id not found", func(t *testing.T) {
		is := is.New(t)
		err := s.SetRepeats("bar", true)
		is.Equal(err, ErrNotFound)
	})
	t.Run("fails to repeat root", func(t *testing.T) {
		is := is.New(t)
		err := s.SetRepeats("bar", true)
		is.True(err != nil)
	})
	t.Run("repeats any direct children of root", func(t *testing.T) {
		is := is.New(t)
		is.Equal(s.Get("foo").Repeats, false)
		err := s.SetRepeats("foo", true)
		is.NoErr(err)
		is.Equal(s.Get("foo").Repeats, true)
	})
	t.Run("cannot repeat any children that are not direct descendants of root", func(t *testing.T) {
		is := is.New(t)
		err := s.SetRepeats("foo1.1", true)
		is.True(err != nil)
		err = s.SetRepeats("foo1.1.1", true)
		is.True(err != nil)
	})
}

func TestStore_Do(t *testing.T) {
	t.Run("throws error if id not found", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		err := s.Do("bar", time.Time{})
		is.Equal(err, ErrNotFound)
	})

	t.Run("completing a task with children completes of all of its children", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		at := time.Time{}.AddDate(5, 6, 7)
		err := s.Do("foo", at)
		is.NoErr(err)
		fooAndChildren := dfs(s.Root().Children[0])
		for _, c := range fooAndChildren {
			doneAt := c.LastDone()
			if doneAt == nil {
				t.Fatal("not done")
			}
			is.Equal(*doneAt, at)
		}
	})

	t.Run("completing all children makes the parent complete", func(t *testing.T) {
		s := setup()
		is := is.New(t)

		is.NoErr(s.Do("foo1", time.Time{}))
		is.NoErr(s.Do("foo2", time.Time{}))
		lastDone := s.Get("foo").LastDone()
		if lastDone == nil {
			t.Fatal("not done")
		}
		is.Equal(time.Time{}, *lastDone)
	})
	t.Run("completing all children makes the parent complete recursively", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		is.NoErr(s.Do("foo2", time.Time{}))
		is.Equal(s.Get("foo").LastDone(), nil)
		is.NoErr(s.Do("foo1.1", time.Time{}))
		lastDone := s.Get("foo").LastDone()
		if lastDone == nil {
			t.Fatal("not done")
		}
		is.Equal(*lastDone, time.Time{})
	})

	t.Run("fails to do an already done task", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		is.True(s.Do("foo", time.Time{}) == nil)
		is.True(s.Do("foo", time.Time{}) != nil)
		is.True(len(s.Get("foo").DoneHistory) == 1)
	})

	t.Run("computes the next repeating date for a repeatable task", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		err := s.Do("daily routine", time.Time{}.AddDate(0, 0, 3))
		is.NoErr(err)
		nextDue := s.Get("daily routine").NextDue()
		if nextDue == nil {
			t.Fatal("expected a new due")
		}
		is.Equal(*nextDue, time.Time{}.AddDate(0, 0, 4))

		err = s.Do("daily routine", time.Time{}.AddDate(0, 0, 8))
		is.NoErr(err)
		nextDue = s.Get("daily routine").NextDue()
		if nextDue == nil {
			t.Fatal("expected a new due")
		}
		is.Equal(*nextDue, time.Time{}.AddDate(0, 0, 9))
	})
	t.Run("cannot complete repeating task IF the deadline hasn't been reached yet", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		err := s.Do("weekly routine", time.Time{}.AddDate(0, 0, 3))
		is.True(err != nil)
	})
}

func TestStore_Delete(t *testing.T) {
	t.Run("fails to delete inexistent task", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		is.True(s.Delete("bar") != nil)
	})
	t.Run("deletes task and all of its children", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		is.True(s.GetChildren("root")[0] == "foo")
		// try deleting twice, second delete SHOULD fail:
		is.NoErr(s.Delete("foo"))
		is.True(s.Delete("foo") != nil)
		is.True(s.Delete("foo1") != nil)
		is.True(s.Delete("foo1.1") != nil)
		t.Run("deletes itself from parent", func(t *testing.T) {
			is := is.New(t)
			is.True(s.GetChildren("root")[0] != "foo")
		})
	})
}

func TestStore_Log(t *testing.T) {
	t.Run("log invalid id fails", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		is.Equal(s.Log("bar", time.Time{}, time.Time{}), ErrNotFound)
	})
	t.Run("logs valid leaf node", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		is.Equal(len(s.Get("foo1.1.1").Logs), 0)
		is.NoErr(s.Log("foo1.1.1", time.Time{}, time.Time{}.Add(time.Hour)))
		is.Equal(len(s.Get("foo1.1.1").Logs), 1)
	})
	t.Run("fails to log when start >= end", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		is.True(s.Log("foo", time.Time{}.Add(time.Hour), time.Time{}) != nil)
	})
	t.Run("fails to log node with children", func(t *testing.T) {
		s := setup()
		is := is.New(t)
		is.True(s.Log("foo", time.Time{}, time.Time{}.Add(time.Hour)) != nil)
	})

}

package date

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestDate_Once(t *testing.T) {
	is := is.New(t)

	absolute := time.Now()
	d := NewOnce(absolute)

	t.Run("In the future", func(t *testing.T) {
		is := is.New(t)
		next := d.Next(time.Time{})
		is.Equal(next, absolute)
	})
}

func TestDate_OnceAYear(t *testing.T) {

	d := NewOnceAYear(21, time.April)
	t.Run("Same day and month", func(t *testing.T) {
		is := is.New(t)
		expected := time.Time{}.AddDate(0, 3, 20)
		next := d.Next(expected)
		expected = expected.AddDate(0, 0, 0)
		is.Equal(expected, next)
	})
	t.Run("Same month", func(t *testing.T) {
		is := is.New(t)
		start := time.Time{}.AddDate(0, 1, 20)
		next := d.Next(start)
		expected := start.AddDate(0, 2, 0)
		is.Equal(expected, next)
	})
	t.Run("Same day", func(t *testing.T) {
		is := is.New(t)
		start := time.Time{}.AddDate(0, 3, 19)
		next := d.Next(start)
		expected := start.AddDate(0, 0, 1)
		is.Equal(expected, next)
	})
}

func TestDate_DayOfTheMonth(t *testing.T) {

	d := NewDayOfTheMonth(21)
	t.Run("After current day", func(t *testing.T) {
		is := is.New(t)
		expected := time.Time{}.AddDate(0, 3, 19)
		next := d.Next(expected)
		expected = expected.AddDate(0, 0, 1)
		is.Equal(expected, next)
	})
	t.Run("Already happened today", func(t *testing.T) {
		is := is.New(t)
		expected := time.Time{}.AddDate(0, 3, 20)
		next := d.Next(expected)
		expected = expected.AddDate(0, 1, 0)
		is.Equal(expected, next)
	})
	t.Run("Already happened this month (before today)", func(t *testing.T) {
		is := is.New(t)
		expected := time.Time{}.AddDate(0, 3, 25)
		next := d.Next(expected)
		expected = expected.AddDate(0, 0, 25)
		is.Equal(expected, next)
	})
}

// TODO: Ideally, we should probably also test for all weekdays
func TestDate_Weekday(t *testing.T) {

	d := NewWeekday(time.Tuesday)
	t.Run("Same time next week", func(t *testing.T) {
		is := is.New(t)
		start := time.Time{}.AddDate(0, 0, 1) // Tuesday
		next := d.Next(start)
		expected := start.AddDate(0, 0, 7)
		is.Equal(expected, next)
	})
	t.Run("In a few days", func(t *testing.T) {
		is := is.New(t)
		start := time.Time{}.AddDate(0, 0, 3) // Thursday
		next := d.Next(start)
		expected := start.AddDate(0, 0, 5)
		is.Equal(expected, next)
	})
}

func TestDate_DayOffset(t *testing.T) {

	t.Run("Positive offset", func(t *testing.T) {
		is := is.New(t)
		start := time.Time{}.AddDate(0, 3, 19)
		next := NewDayOffset(10).Next(start)
		expected := start.AddDate(0, 0, 10)
		is.Equal(expected, next)
	})
	t.Run("Negative offset", func(t *testing.T) {
		is := is.New(t)
		start := time.Time{}.AddDate(0, 3, 19)
		next := NewDayOffset(-50).Next(start)
		expected := start.AddDate(0, 0, -50)
		is.Equal(expected, next)
	})
	t.Run("Same day", func(t *testing.T) {
		is := is.New(t)
		start := time.Time{}.AddDate(0, 3, 19)
		next := NewDayOffset(0).Next(start)
		is.Equal(next, start)
	})
}

package date

import (
	"time"
)

type PeriodType int

const (
	Once PeriodType = iota
	OnceAYear
	DayOfTheMonth
	Weekday
	DayOffset
)

type dayMonth struct {
	Day   int
	Month time.Month
}

// RepeatableDate stores information about a absolute/relative date
// It uses interface{} because it needs to be serialised to JSON
// Ideally this struct would be an interface { Next(time.Time) *time.Time }
type RepeatableDate struct {
	Type  PeriodType
	Value interface{}
}

func NewOnce(t time.Time) RepeatableDate {
	return RepeatableDate{Type: Once, Value: t}
}

func NewOnceAYear(day int, month time.Month) RepeatableDate {
	return RepeatableDate{Type: OnceAYear, Value: dayMonth{Day: day, Month: month}}
}

func NewDayOfTheMonth(day int) RepeatableDate {
	return RepeatableDate{Type: DayOfTheMonth, Value: day}
}

func NewWeekday(weekday time.Weekday) RepeatableDate {
	return RepeatableDate{Type: Weekday, Value: weekday}
}

func NewDayOffset(days int) RepeatableDate {
	return RepeatableDate{Type: DayOffset, Value: days}
}

// returns a date after the provided date
// errors if cannot compute a new date
func (d RepeatableDate) Next(t time.Time) *time.Time {
	switch d.Type {
	case Once:
		newDate := d.Value.(time.Time)
		if newDate.Before(t) {
			return nil
		}
		return &newDate
	case OnceAYear:
		dm := d.Value.(dayMonth)
		months := int(dm.Month) - int(t.Month())
		days := dm.Day - t.Day()
		years := 0
		if months < 0 || days < 0 || months == 0 && days == 0 {
			years = 1
		}
		next := t.AddDate(years, months, days)
		return &next
	case DayOfTheMonth:
		nth := d.Value.(int)
		months := 0
		days := nth - t.Day()
		if days <= 0 {
			months = 1
		}
		next := t.AddDate(0, months, days)
		return &next
	case Weekday:
		w := d.Value.(time.Weekday)
		days := int(w - t.Weekday())
		if days <= 0 {
			days += 7
		}
		next := t.AddDate(0, 0, days)
		return &next
	case DayOffset:
		days := d.Value.(int)
		if days == 0 {
			return nil
		}
		next := t.AddDate(0, 0, days)
		return &next
	default:
		panic("unimplemented")
	}
}

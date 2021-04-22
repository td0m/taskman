package date

import (
	"encoding/json"
	"log"
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

// TODO: test for unmarshal
func (r *RepeatableDate) UnmarshalJSON(bs []byte) error {
	type alias RepeatableDate
	var out alias
	if err := json.Unmarshal(bs, &out); err != nil {
		return err
	}
	r.Type = out.Type
	switch r.Type {
	case Once:
		var err error
		r.Value, err = time.Parse(time.RFC3339, out.Value.(string))
		return err
	case DayOffset, DayOfTheMonth:
		r.Value = int(out.Value.(float64))
		return nil
	case Weekday:
		r.Value = time.Weekday(out.Value.(float64))
		return nil
	case OnceAYear:
		m := out.Value.(map[string]interface{})
		r.Value = dayMonth{Day: int(m["Day"].(float64)), Month: time.Month(m["Month"].(float64))}
		return nil
	}
	switch v := out.Value.(type) {
	default:
		log.Panicf("%+v", v)
	}
	return nil
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
func (d RepeatableDate) Next(t time.Time) time.Time {
	switch d.Type {
	case Once:
		newDate := d.Value.(time.Time)
		return newDate
	case OnceAYear:
		dm := d.Value.(dayMonth)
		months := int(dm.Month) - int(t.Month())
		days := dm.Day - t.Day()
		years := 0
		if months < 0 || days < 0 || months == 0 && days == 0 {
			years = 1
		}
		return t.AddDate(years, months, days)
	case DayOfTheMonth:
		nth := d.Value.(int)
		months := 0
		days := nth - t.Day()
		if days <= 0 {
			months = 1
		}
		return t.AddDate(0, months, days)
	case Weekday:
		w := d.Value.(time.Weekday)
		days := int(w - t.Weekday())
		if days <= 0 {
			days += 7
		}
		return t.AddDate(0, 0, days)
	case DayOffset:
		days := d.Value.(int)
		next := t.AddDate(0, 0, days)
		return next
	default:
		panic("unimplemented")
	}
}

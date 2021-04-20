package date

import (
	"reflect"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	today := time.Now().Truncate(time.Hour * 24)
	tests := []struct {
		name    string
		args    []string
		want    RepeatableDate
		wantErr bool
	}{
		{"Case Insensitive", []string{"ToDAY"}, RepeatableDate{Type: Once, Value: today}, false},
		{"today", []string{"today", "tod"}, RepeatableDate{Type: Once, Value: today}, false},
		{"tomorrow", []string{"tomorrow", "tom", "1", "+1", "in 1 day", "1d", "1day", "1 day"}, RepeatableDate{Type: DayOffset, Value: 1}, false},
		{"yesterday", []string{"yday", "yesterday", "1 day ago", "1d ago", "-1"}, RepeatableDate{Type: DayOffset, Value: -1}, false},
		{"7 days", []string{"7 day", "7 days", "1 week", "7"}, RepeatableDate{Type: DayOffset, Value: 7}, false},
		{"1 month", []string{"1 month", "1m"}, RepeatableDate{Type: DayOffset, Value: 30}, false},
		{"1 year", []string{"1y", "in 1 year"}, RepeatableDate{Type: DayOffset, Value: 365}, false},

		{"absolute", []string{"20/04/21", "20/04/2021", "20 April 2021", "20 Apr 2021"}, RepeatableDate{Type: Once, Value: time.Time{}.AddDate(2020, 3, 19)}, false},
		{"monday", []string{"mon", "monday"}, RepeatableDate{Type: Weekday, Value: time.Monday}, false},
		{"tuesday", []string{"tue", "tuesday"}, RepeatableDate{Type: Weekday, Value: time.Tuesday}, false},
		{"wednesday", []string{"wed", "wednesday"}, RepeatableDate{Type: Weekday, Value: time.Wednesday}, false},
		{"thursday", []string{"thu", "thursday"}, RepeatableDate{Type: Weekday, Value: time.Thursday}, false},
		{"friday", []string{"fri", "friday"}, RepeatableDate{Type: Weekday, Value: time.Friday}, false},
		{"saturday", []string{"sat", "saturday"}, RepeatableDate{Type: Weekday, Value: time.Saturday}, false},
		{"sunday", []string{"sun", "sunday"}, RepeatableDate{Type: Weekday, Value: time.Sunday}, false},

		// TODO: test that 0, negative dates, invalid postfixes, and too high dates (like 40th) don't work
		{"1st", []string{"1st"}, RepeatableDate{Type: DayOfTheMonth, Value: 1}, false},
		{"2nd", []string{"2nd"}, RepeatableDate{Type: DayOfTheMonth, Value: 2}, false},
		{"3rd", []string{"3rd"}, RepeatableDate{Type: DayOfTheMonth, Value: 3}, false},
		{"4th", []string{"4th"}, RepeatableDate{Type: DayOfTheMonth, Value: 4}, false},
		{"18th", []string{"18th"}, RepeatableDate{Type: DayOfTheMonth, Value: 18}, false},
		{"10th", []string{"10th"}, RepeatableDate{Type: DayOfTheMonth, Value: 10}, false},

		{"1st Jan", []string{"1st Jan", "1st January"}, RepeatableDate{Type: OnceAYear, Value: dayMonth{1, 1}}, false},
		{"1st Feb", []string{"1st Feb", "1st February"}, RepeatableDate{Type: OnceAYear, Value: dayMonth{1, 2}}, false},
		{"1st Dec", []string{"1st Dec", "1st December"}, RepeatableDate{Type: OnceAYear, Value: dayMonth{1, 12}}, false},

		// TODO: Add more test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, arg := range tt.args {
				got, err := ParseDate(arg)
				if (err != nil) != tt.wantErr {
					t.Errorf("ParseDate(%s) error = %v, wantErr %v", arg, err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParseDate(%s) = %v, want %v", arg, got, tt.want)
				}
			}
		})
	}
}

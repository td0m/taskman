package dateinput

import (
	"testing"
	"time"
)

func Test_parseRelative(t *testing.T) {
	var (
		day = time.Duration(time.Hour * 24)
	)
	now := time.Now()
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"in1", day, false},
		{"in 1", day, false},
		{"1", day, false},
		{"in 11", day * 11, false},
		{"in 231", day * 231, false},
		{"in 1 day", day, false},
		{"in 1 days", day, false},
		{"in 1 week", day * 7, false},
		{"in 1 month", day * 30, false},
		{"in 2 year", day * 365 * 2, false},
		{"in 1w", day * 7, false},
		{"in 1wek", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseRelative(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRelative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("%s\ngot:  %v\nwant: %v", tt.input, now.Add(got).Format("02-01-2006"), now.Add(tt.want).Format("02-01-2006"))
			}
		})
	}
}

func Test_parseAbsolute(t *testing.T) {
	now, _ := time.Parse("02-01-2006", "01-02-2006")
	tests := []struct {
		input   string
		want    time.Time
		wantErr bool
	}{
		{"21-04", now.AddDate(0, 2, 20), false},
		{"21", now.AddDate(0, 0, 20), false},
		{"21-04-06", now.AddDate(0, 2, 20), false},
		{"feb 21", now.AddDate(0, 0, 20), false},
		{"february 21", now.AddDate(0, 0, 20), false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseAbsolute(tt.input, now)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRelative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("%s\ngot:  %v\nwant: %v", tt.input, got.Format("02-01-2006"), tt.want.Format("02-01-2006"))
			}
		})
	}
}

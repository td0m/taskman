package dateinput

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type multiplier struct {
	key   string
	value int
}

var multipliers = []multiplier{
	{"days", 1},
	{"weeks", 7},
	{"months", 30},
	{"years", 365},
}

func parseRelative(s string) (time.Duration, error) {
	s = strings.TrimPrefix(s, "in")
	s = strings.TrimSpace(s)
	var n int
	// parse quantity
	{
		i := 0
		for {
			if i >= len(s) {
				break
			}
			var err error
			n1, err := strconv.Atoi(s[:i+1])
			// first one can not fail
			if err != nil {
				if i == 0 {
					return 0, err
				} else {
					break
				}
			}
			n = n1
			i++
		}
		s = strings.TrimSpace(s[i:])
	}

	multiplier := 1
	if len(s) > 0 {
		multiplier = 0
		for _, m := range multipliers {
			end := min(len(m.key), len(s))
			if m.key[:end] == s {
				multiplier = m.value
				break
			}
		}
		if multiplier == 0 {
			return 0, errors.New("unexpected postfix")
		}
	}

	return time.Duration(n*multiplier) * time.Hour * 24, nil
}

func parseAnyFormat(s string) (time.Time, error) {
	for _, fmt := range formats {
		t, err := time.Parse(fmt, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("format not found")
}

func parseAbsolute(s string, now time.Time) (time.Time, error) {
	t, err := parseAnyFormat(s)
	if err != nil {
		return t, err
	}
	year, month := 0, 1
	if t.Year() == 0 {
		year = now.Year()
		if t.Month() == 1 {
			month = int(now.Month())
		}
	}
	return t.AddDate(year, month-1, 0), nil
}

var formats = []string{
	"_2",
	"_2/01",
	"_2/01/06",
	"_2/01/2006",
	"_2-01",
	"_2-01-06",
	"_2-01-2006",
	"Jan _2",
	"Jan _2 06",
	"Jan _2 2006",
	"January _2",
	"January _2 06",
	"January _2 2006",
	"_2 Jan",
	"_2 Jan 06",
	"_2 Jan 2006",
	"_2 January",
	"_2 January 06",
	"_2 January 2006",
}

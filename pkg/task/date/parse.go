package date

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var ErrParsing = errors.New("error parsing date")

func ParseDate(s string) (RepeatableDate, error) {
	s = strings.ToLower(s)
	// today is a special case
	// it cannot be relative, since it will always give the same day
	// TODO: consider changing recursive mechanism to make sure this doesn't happen
	if s == "today" || s == "tod" || s == "now" {
		return NewOnce(time.Now().Truncate(time.Hour * 24)), nil
	}
	if s == "tomorrow" || s == "tom" {
		return NewDayOffset(1), nil
	}
	if s == "yesterday" || s == "yday" {
		return NewDayOffset(-1), nil
	}
	wkd, err := parseWeekday(s)
	if err == nil {
		return wkd, err
	}
	rel, err := parseOffset(s)
	if err == nil {
		return rel, nil
	}
	abs, err := parseAbsolute(s)
	if err == nil {
		return abs, nil
	}

	return RepeatableDate{}, errors.New("no format matches")
}

func parseAnyTimeFormat(s string, formats []string) (time.Time, error) {
	for _, fmt := range formats {
		t, err := time.Parse(fmt, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("format not found")
}

func parseAbsolute(s string) (RepeatableDate, error) {
	t, err := parseAnyTimeFormat(s, absoluteFormats)
	if err != nil {
		return RepeatableDate{}, err
	}
	return NewOnce(t), nil
}

var absoluteFormats = []string{
	"_2/01/06",
	"_2/01/2006",
	"_2 Jan 2006",
	"_2 January 2006",
}

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

// TODO: add month/year repeatabledates
// this will account for when month != 31 days and leap years
func parseOffset(s string) (RepeatableDate, error) {
	s = strings.TrimPrefix(s, "in")
	s = strings.TrimSpace(s)
	var (
		n        int
		negative bool
	)
	if len(s) >= 1 {
		if s[0] == '-' {
			negative = true
			s = s[1:]
		}
		if s[0] == '+' {
			s = s[1:]
		}
	}
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
					return RepeatableDate{}, err
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
		endOfWord := len(s)
		for i, c := range s {
			if c == ' ' {
				endOfWord = i
				break
			}
		}
		for _, m := range multipliers {
			end := min(len(m.key), endOfWord)
			if m.key[:end] == s[:end] {
				multiplier = m.value
				s = s[end:]
				break
			}
		}
		s = strings.TrimSpace(s)
		if s == "ago" {
			negative = true
		}
		if multiplier == 0 {
			return RepeatableDate{}, errors.New("invalid suffix, expected 'days', 'months', 'weeks', or 'years'")
		}
	}

	if negative {
		n *= -1
	}

	return NewDayOffset(n * multiplier), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseWeekday(s string) (RepeatableDate, error) {
	for i := time.Sunday; i <= time.Saturday; i++ {
		fmt := strings.ToLower(i.String())
		if s == fmt || s == fmt[:3] {
			return NewWeekday(i), nil
		}
	}
	return RepeatableDate{}, errors.New("invalid weekday")
}

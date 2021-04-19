package date

import (
	"errors"
	"strings"
	"time"
)

var ErrParsing = errors.New("error parsing date")

func ParseDate(s string) (RepeatableDate, error) {
	s = strings.ToLower(s)
	if s == "today" || s == "now" {
		return NewOnce(time.Now().Truncate(time.Hour * 24)), nil
	}

	return RepeatableDate{}, nil
}

package timeutil

import (
	"errors"
	"strings"
	"time"
)

var errNoTimeSpecified = errors.New("No time specified")

const (
	formatDate      = "2006-01-02"
	formatShortDate = "01-02"
)

// ParseExpr is a convenience interface to [ParseExprRef] which provides
// [time.Now] as the reference time. It's usually the one you want.
func ParseExpr(s string) (time.Time, error) {
	return ParseExprRef(s, time.Now())
}

// ParseExprRef parses a time expression and returns the point in time that
// it represents. Many expression refer to relative time, which is evaluated
// relative to the provided reference time.
//
// This function supports a variety of inputs:
//
//   - The special constants: "today", "yesterday", and "tomorrow", which refers
//     to midnight on those days, relative to the reference time;
//
//   - The special constant: "now", which refers to the reference time, which
//     is simply returned;
//
//   - A relative time adjustment, in the form: "(+|-)duration", where
//     "duration" is a duration (as implemented in this package) relative to the
//     reference time. For example, the expression "-10d" refers to the point in
//     time 10 days ago at the same time as this function is invoked;
//
//   - A date expressed as the day and month, which is assumed to be in the
//     reference year; for example "11-14" refers to midnight on November 14th of
//     the year of the reference time;
//
//   - A date expressed as the day, month, and year without a time, which
//     refers to midnight on that date.
//
// Any other input, including an empty string is an error.
func ParseExprRef(s string, ref time.Time) (time.Time, error) {
	v := strings.TrimSpace(s)
	if v == "" {
		return time.Time{}, errNoTimeSpecified
	}
	switch v { // constants
	case "today":
		return ref.Truncate(time.Hour * 24), nil
	case "yesterday":
		return ref.Truncate(time.Hour*24).AddDate(0, 0, -1), nil
	case "tomorrow":
		return ref.Truncate(time.Hour*24).AddDate(0, 0, 1), nil
	case "now":
		return ref, nil
	}
	if f := v[0]; f == '+' || f == '-' { // time must have at least 1 index since it's not ""
		d, err := ParseDuration(v)
		if err != nil {
			return time.Time{}, err
		}
		return ref.Add(d), nil
	} else if len(v) == len(formatShortDate) {
		t, err := time.Parse(formatDate, ref.Format("2006")+"-"+v) // assume current year
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	} else if len(v) == len(formatDate) {
		t, err := time.Parse(formatDate, v)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	} else {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}
}

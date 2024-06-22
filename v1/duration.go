package timeutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(FormatDuration(time.Duration(d)))
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, err := ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

const (
	lowerhex  = "0123456789abcdef"
	runeSelf  = 0x80
	runeError = '\uFFFD'
)

var unitMap = map[string]uint64{
	"ns": uint64(time.Nanosecond),
	"us": uint64(time.Microsecond),
	"µs": uint64(time.Microsecond), // U+00B5 = micro symbol
	"μs": uint64(time.Microsecond), // U+03BC = Greek letter mu
	"ms": uint64(time.Millisecond),
	"s":  uint64(time.Second),
	"m":  uint64(time.Minute),
	"h":  uint64(time.Hour),
	"d":  uint64(time.Hour * 24),
	"w":  uint64(time.Hour * 24 * 7),
}

// ParseDuration parses a duration string. This works the same as
// time.ParseDuration except that it supports additional approximated
// units:
//   - d: days (defined as 24 hours)
//   - w: weeks (defined as 7 days)
func ParseDuration(s string) (time.Duration, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d uint64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, errors.New("time: invalid duration " + quote(orig))
	}
	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, errors.New("time: invalid duration " + quote(orig))
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, errors.New("time: invalid duration " + quote(orig))
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, errors.New("time: invalid duration " + quote(orig))
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, errors.New("time: missing unit in duration " + quote(orig))
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, errors.New("time: unknown unit " + quote(u) + " in duration " + quote(orig))
		}
		if v > 1<<63/unit {
			// overflow
			return 0, errors.New("time: invalid duration " + quote(orig))
		}
		v *= unit
		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += uint64(float64(f) * (float64(unit) / scale))
			if v > 1<<63 {
				// overflow
				return 0, errors.New("time: invalid duration " + quote(orig))
			}
		}
		d += v
		if d > 1<<63 {
			return 0, errors.New("time: invalid duration " + quote(orig))
		}
	}
	if neg {
		return -time.Duration(d), nil
	}
	if d > 1<<63-1 {
		return 0, errors.New("time: invalid duration " + quote(orig))
	}
	return time.Duration(d), nil
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x uint64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x uint64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + uint64(c) - '0'
		if y > 1<<63 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

func quote(s string) string {
	buf := make([]byte, 1, len(s)+2) // slice will be at least len(s) + quotes
	buf[0] = '"'
	for i, c := range s {
		if c >= runeSelf || c < ' ' {
			// This means you are asking us to parse a time.Duration or
			// time.Location with unprintable or non-ASCII characters in it.
			// We don't expect to hit this case very often. We could try to
			// reproduce strconv.Quote's behavior with full fidelity but
			// given how rarely we expect to hit these edge cases, speed and
			// conciseness are better.
			var width int
			if c == runeError {
				width = 1
				if i+2 < len(s) && s[i:i+3] == string(runeError) {
					width = 3
				}
			} else {
				width = len(string(c))
			}
			for j := 0; j < width; j++ {
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[i+j]>>4])
				buf = append(buf, lowerhex[s[i+j]&0xF])
			}
		} else {
			if c == '"' || c == '\\' {
				buf = append(buf, '\\')
			}
			buf = append(buf, string(c)...)
		}
	}
	buf = append(buf, '"')
	return string(buf)
}

const day = time.Hour * 24

func formatHigh(d time.Duration) string {
	var v time.Duration
	var f string

	d = d / time.Second
	v = d % 60
	if v != 0 {
		f = fmt.Sprintf("%ds", v) + f
	}

	d /= 60
	v = d % 60
	if v != 0 {
		f = fmt.Sprintf("%dm", v) + f
	}

	d /= 60
	v = d % 24
	if v != 0 {
		f = fmt.Sprintf("%dh", v) + f
	}

	d /= 24
	if d > 0 {
		f = fmt.Sprintf("%dd", d) + f
	}

	return f
}

func formatLow(d time.Duration) string {
	var v time.Duration
	var f string

	d = d % time.Second
	v = d % 1000
	if v != 0 {
		f = fmt.Sprintf("%dns", v) + f
	}

	d /= 1000
	v = d % 1000
	if v != 0 {
		f = fmt.Sprintf("%dµs", v) + f
	}

	d /= 1000
	if d > 0 {
		f = fmt.Sprintf("%dms", d) + f
	}

	return f
}

func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	} else {
		return formatHigh(d) + formatLow(d)
	}
}

func FormatSimplifiedDuration(d time.Duration) string {
	switch {
	case d > time.Hour*24:
		return fmt.Sprintf("%dd %dh", d.Truncate(time.Hour*24)/(time.Hour*24), (d%(time.Hour*24))/time.Hour)
	case d > time.Hour:
		return fmt.Sprintf("%dh", d.Truncate(time.Hour)/time.Hour)
	case d > time.Minute:
		return fmt.Sprintf("%dm", d.Truncate(time.Minute)/time.Minute)
	case d > time.Second:
		return fmt.Sprintf("%ds", d.Truncate(time.Second)/time.Second)
	case d > time.Millisecond:
		return fmt.Sprintf("%dms", d.Truncate(time.Millisecond)/time.Millisecond)
	case d > time.Microsecond:
		return fmt.Sprintf("%dµs", d.Truncate(time.Microsecond)/time.Microsecond)
	default:
		return fmt.Sprintf("%dns", d)
	}
}

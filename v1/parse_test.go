package timeutil

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseExpr(t *testing.T) {
	ref := time.Date(2024, 11, 14, 18, 17, 0, 0, time.UTC)
	tests := []struct {
		Ref    time.Time
		Expr   string
		Expect time.Time
		Err    func(error) error
	}{
		{
			Ref:    ref,
			Expr:   "now",
			Expect: ref,
		},
		{
			Ref:    ref,
			Expr:   "yesterday",
			Expect: ref.Truncate(time.Hour*24).AddDate(0, 0, -1),
		},
		{
			Ref:    ref,
			Expr:   "tomorrow",
			Expect: ref.Truncate(time.Hour*24).AddDate(0, 0, 1),
		},
		{
			Ref:    ref,
			Expr:   "today",
			Expect: ref.Truncate(time.Hour * 24),
		},
		{
			Ref:    ref,
			Expr:   "-1h",
			Expect: ref.Add(-time.Hour),
		},
		{
			Ref:    ref,
			Expr:   "+1d",
			Expect: ref.Add(time.Hour * 24),
		},
		{
			Ref:    ref,
			Expr:   "05-01",
			Expect: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Ref:    ref,
			Expr:   "2021-05-01",
			Expect: time.Date(2021, 5, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Ref:  ref,
			Expr: "",
			Err: func(err error) error {
				if errors.Is(err, errNoTimeSpecified) {
					return nil
				} else {
					return err
				}
			},
		},
		{
			Ref:  ref,
			Expr: "???",
			Err: func(err error) error {
				if err != nil {
					return nil
				} else {
					return errors.New("Expected an error")
				}
			},
		},
	}
	for i, test := range tests {
		v, err := ParseExprRef(test.Expr, test.Ref)
		if test.Err != nil {
			assert.NoError(t, test.Err(err), "#%d", i)
		} else if assert.NoError(t, err, "#%d", i) {
			assert.Equal(t, test.Expect, v, "#%d", i)
		}
	}
}

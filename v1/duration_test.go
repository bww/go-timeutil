package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "366d1h1m1s1ms1µs1ns", FormatDuration((day*366)+time.Hour+time.Minute+time.Second+time.Millisecond+time.Microsecond+time.Nanosecond))
	assert.Equal(t, "8d8h8m8s8ms8µs8ns", FormatDuration((day*8)+(time.Hour*8)+(time.Minute*8)+(time.Second*8)+(time.Millisecond*8)+(time.Microsecond*8)+(time.Nanosecond*8)))
	assert.Equal(t, "8h8m8s8ms8µs8ns", FormatDuration((time.Hour*8)+(time.Minute*8)+(time.Second*8)+(time.Millisecond*8)+(time.Microsecond*8)+(time.Nanosecond*8)))
	assert.Equal(t, "8m8s8ms8µs8ns", FormatDuration((time.Minute*8)+(time.Second*8)+(time.Millisecond*8)+(time.Microsecond*8)+(time.Nanosecond*8)))
	assert.Equal(t, "8s8ms8µs8ns", FormatDuration((time.Second*8)+(time.Millisecond*8)+(time.Microsecond*8)+(time.Nanosecond*8)))
	assert.Equal(t, "8ms8µs8ns", FormatDuration((time.Millisecond*8)+(time.Microsecond*8)+(time.Nanosecond*8)))
	assert.Equal(t, "8µs8ns", FormatDuration((time.Microsecond*8)+(time.Nanosecond*8)))
	assert.Equal(t, "8ns", FormatDuration((time.Nanosecond * 8)))
	assert.Equal(t, "8d", FormatDuration(day*8))
	assert.Equal(t, "8h", FormatDuration(time.Hour*8))
	assert.Equal(t, "8m", FormatDuration(time.Minute*8))
	assert.Equal(t, "8s", FormatDuration(time.Second*8))
	assert.Equal(t, "800ms", FormatDuration(time.Millisecond*800))
	assert.Equal(t, "800µs", FormatDuration(time.Microsecond*800))
	assert.Equal(t, "800ns", FormatDuration(time.Nanosecond*800))
	assert.Equal(t, "1ns", FormatDuration(time.Nanosecond))
}

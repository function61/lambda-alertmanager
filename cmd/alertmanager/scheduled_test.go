package main

import (
	"fmt"
	"github.com/function61/gokit/assert"
	"testing"
	"time"
)

func TestParseTtlSpec(t *testing.T) {
	now := time.Date(2019, 9, 7, 12, 0, 0, 0, time.UTC)

	tcs := []struct {
		input  string
		output string
	}{
		{
			"+24h",
			"2019-09-08T12:00:00Z",
		},
		{
			"+1h",
			"2019-09-07T13:00:00Z",
		},
		{
			"+1d@18:00",
			"2019-09-08T18:00:00Z",
		},
		{
			"+14d@10:00",
			"2019-09-21T10:00:00Z",
		},
		{
			"2019-09-10T01:13:00Z",
			"2019-09-10T01:13:00Z",
		},
		{
			"foobar",
			"error: not in RFC3339: foobar",
		},
		{
			"+12x",
			"error: duration in bad format",
		},
		{
			"+1d@18:0x",
			"error: duration in bad format",
		},
	}

	for _, tc := range tcs {
		tc := tc // pin
		t.Run(tc.input, func(t *testing.T) {
			ttl, err := parseTtlSpec(tc.input, now)
			var actual string
			if err != nil {
				actual = fmt.Sprintf("error: %v", err)
			} else {
				actual = ttl.Format(time.RFC3339Nano)
			}

			assert.EqualString(t, actual, tc.output)
		})
	}
}

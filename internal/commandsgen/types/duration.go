// NOTE: this file is embedded inside the generated commands.gen.go output

package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var reDays = regexp.MustCompile(`(\d+(\.\d*)?|(\.\d+))d`)

type Duration time.Duration

// ParseDuration is like time.ParseDuration, but supports unit "d" for days
// (always interpreted as exactly 24 hours).
func ParseDuration(s string) (time.Duration, error) {
	s = reDays.ReplaceAllStringFunc(s, func(v string) string {
		fv, err := strconv.ParseFloat(strings.TrimSuffix(v, "d"), 64)
		if err != nil {
			return v // will cause time.ParseDuration to return an error
		}
		return fmt.Sprintf("%fh", 24*fv)
	})
	return time.ParseDuration(s)
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d *Duration) String() string {
	return d.Duration().String()
}

func (d *Duration) Set(s string) error {
	p, err := ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(p)
	return nil
}

func (d *Duration) Type() string {
	return "duration"
}

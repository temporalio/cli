package commandsgen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var reDays = regexp.MustCompile(`(\d+(\.\d*)?|(\.\d+))d`)

// parseDuration is like time.ParseDuration, but supports unit "d" for days
// (always interpreted as exactly 24 hours).
func parseDuration(s string) (time.Duration, error) {
	s = reDays.ReplaceAllStringFunc(s, func(v string) string {
		fv, err := strconv.ParseFloat(strings.TrimSuffix(v, "d"), 64)
		if err != nil {
			return v // will cause time.ParseDuration to return an error
		}
		return fmt.Sprintf("%fh", 24*fv)
	})
	return time.ParseDuration(s)
}

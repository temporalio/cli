package temporalcli

import (
	"time"

	"go.temporal.io/server/common/primitives/timestamp"
)

type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d *Duration) String() string {
	return d.Duration().String()
}

func (d *Duration) Set(s string) error {
	p, err := timestamp.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(p)
	return nil
}

func (d *Duration) Type() string {
	return "duration"
}

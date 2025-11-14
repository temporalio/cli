package temporalcli

import "time"

type Timestamp time.Time

func (t Timestamp) Time() time.Time {
	return time.Time(t)
}

func (t *Timestamp) String() string {
	return t.Time().Format(time.RFC3339)
}

func (t *Timestamp) Set(s string) error {
	p, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*t = Timestamp(p)
	return nil
}

func (t *Timestamp) Type() string {
	return "timestamp"
}

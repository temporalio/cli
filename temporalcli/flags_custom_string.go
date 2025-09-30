package temporalcli

import (
	"fmt"

	"github.com/spf13/pflag"
)

// customStringFlag implements pflag.Value while storing the raw string into the
// provided target. Its Type() reports a user-facing display type (e.g., "float")
// so help/usage can show the expected format even though parsing is handled later.
type customStringFlag struct {
	target      *string
	displayType string
}

func newCustomStringFlag(target *string, displayType string) pflag.Value {
	ret := &customStringFlag{
		target:      target,
		displayType: displayType,
	}
	return ret
}

func (v *customStringFlag) Set(s string) error {
	if v == nil || v.target == nil {
		return fmt.Errorf("internal error: customStringFlag target is nil")
	}
	*v.target = s
	return nil
}

func (v *customStringFlag) String() string {
	if v == nil || v.target == nil || *v.target == "" {
		return ""
	}
	return *v.target
}

func (v *customStringFlag) Type() string {
	return v.displayType
}

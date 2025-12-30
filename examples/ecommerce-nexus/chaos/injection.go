package chaos

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds chaos injection configuration
type Config struct {
	// Rate is the probability of failure (0.0 - 1.0)
	Rate float64
	// Types of failures to inject: timeout, error, panic
	Types []string
	// Services to target (empty = all)
	Services []string
}

// DefaultConfig returns a disabled chaos config
func DefaultConfig() *Config {
	return &Config{
		Rate:     0.0,
		Types:    []string{"error"},
		Services: []string{},
	}
}

// FromEnv loads chaos config from environment variables
func FromEnv() *Config {
	cfg := DefaultConfig()

	if rate := os.Getenv("CHAOS_RATE"); rate != "" {
		if r, err := strconv.ParseFloat(rate, 64); err == nil {
			cfg.Rate = r
		}
	}

	if types := os.Getenv("CHAOS_TYPES"); types != "" {
		cfg.Types = strings.Split(types, ",")
	}

	if services := os.Getenv("CHAOS_SERVICES"); services != "" {
		cfg.Services = strings.Split(services, ",")
	}

	return cfg
}

// Injector handles chaos injection
type Injector struct {
	cfg *Config
	rng *rand.Rand
}

// NewInjector creates a new chaos injector
func NewInjector(cfg *Config) *Injector {
	return &Injector{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// MaybeInject potentially injects a failure based on configuration
func (i *Injector) MaybeInject(ctx context.Context, service, operation string) error {
	// Check if this service should be targeted
	if len(i.cfg.Services) > 0 {
		found := false
		for _, s := range i.cfg.Services {
			if s == service {
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	// Check probability
	if i.rng.Float64() > i.cfg.Rate {
		return nil
	}

	// Pick a failure type
	failType := i.cfg.Types[i.rng.Intn(len(i.cfg.Types))]

	switch failType {
	case "timeout":
		// Simulate timeout by sleeping longer than typical timeouts
		select {
		case <-time.After(30 * time.Second):
			return errors.New("chaos: simulated timeout")
		case <-ctx.Done():
			return ctx.Err()
		}
	case "panic":
		panic(fmt.Sprintf("chaos: simulated panic in %s.%s", service, operation))
	case "error":
		fallthrough
	default:
		return fmt.Errorf("chaos: injected failure in %s.%s", service, operation)
	}

	return nil
}

// Enabled returns true if chaos injection is enabled
func (i *Injector) Enabled() bool {
	return i.cfg.Rate > 0
}

// Global injector for convenience
var global = NewInjector(DefaultConfig())

// Init initializes the global chaos injector from environment
func Init() {
	global = NewInjector(FromEnv())
}

// MaybeInject uses the global injector
func MaybeInject(ctx context.Context, service, operation string) error {
	return global.MaybeInject(ctx, service, operation)
}

// Enabled checks if global chaos is enabled
func Enabled() bool {
	return global.Enabled()
}

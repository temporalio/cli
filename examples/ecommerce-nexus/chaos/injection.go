package chaos

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FailureType defines the type of failure to inject
type FailureType string

const (
	FailureTypeError      FailureType = "error"
	FailureTypeTimeout    FailureType = "timeout"
	FailureTypeLatency    FailureType = "latency"
	FailureTypePanic      FailureType = "panic"
	FailureTypePayment    FailureType = "payment"
	FailureTypeFraud      FailureType = "fraud"
	FailureTypeInventory  FailureType = "inventory"
	FailureTypeShipping   FailureType = "shipping"
	FailureTypeValidation FailureType = "validation"
)

// RealisticErrors maps failure types to realistic error messages
var RealisticErrors = map[FailureType][]string{
	FailureTypePayment: {
		"payment gateway timeout",
		"card declined: insufficient funds",
		"payment processor unavailable",
		"transaction limit exceeded",
		"invalid payment method",
	},
	FailureTypeFraud: {
		"fraud detected: velocity check failed",
		"fraud detected: high risk transaction",
		"fraud detected: suspicious IP address",
		"fraud detected: card used in multiple countries",
	},
	FailureTypeInventory: {
		"item out of stock",
		"warehouse connection timeout",
		"inventory reservation failed",
		"SKU not found in catalog",
	},
	FailureTypeShipping: {
		"carrier API unavailable",
		"invalid shipping address",
		"no carriers available for destination",
		"shipping rate calculation failed",
	},
	FailureTypeValidation: {
		"invalid order: missing required fields",
		"invalid order: quantity exceeds limit",
		"invalid order: customer not verified",
	},
}

// Config holds chaos injection configuration
type Config struct {
	// Rate is the probability of failure (0.0 - 1.0)
	Rate float64
	// Types of failures to inject
	Types []FailureType
	// Services to target (empty = all)
	Services []string
	// MinLatencyMs for latency injection
	MinLatencyMs int
	// MaxLatencyMs for latency injection
	MaxLatencyMs int
	// CascadeRate is the probability that one failure triggers another
	CascadeRate float64
}

// DefaultConfig returns a disabled chaos config
func DefaultConfig() *Config {
	return &Config{
		Rate:         0.0,
		Types:        []FailureType{FailureTypeError},
		Services:     []string{},
		MinLatencyMs: 100,
		MaxLatencyMs: 2000,
		CascadeRate:  0.0,
	}
}

// ProductionChaosConfig returns a config suitable for production simulation
func ProductionChaosConfig() *Config {
	return &Config{
		Rate: 0.15, // 15% of operations fail
		Types: []FailureType{
			FailureTypePayment,
			FailureTypeFraud,
			FailureTypeInventory,
			FailureTypeShipping,
			FailureTypeLatency,
		},
		Services:     []string{},
		MinLatencyMs: 500,
		MaxLatencyMs: 3000,
		CascadeRate:  0.1, // 10% chance of cascading failure
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
		typeStrs := strings.Split(types, ",")
		cfg.Types = make([]FailureType, len(typeStrs))
		for i, t := range typeStrs {
			cfg.Types[i] = FailureType(strings.TrimSpace(t))
		}
	}

	if services := os.Getenv("CHAOS_SERVICES"); services != "" {
		cfg.Services = strings.Split(services, ",")
	}

	if minLatency := os.Getenv("CHAOS_MIN_LATENCY_MS"); minLatency != "" {
		if ms, err := strconv.Atoi(minLatency); err == nil {
			cfg.MinLatencyMs = ms
		}
	}

	if maxLatency := os.Getenv("CHAOS_MAX_LATENCY_MS"); maxLatency != "" {
		if ms, err := strconv.Atoi(maxLatency); err == nil {
			cfg.MaxLatencyMs = ms
		}
	}

	if cascade := os.Getenv("CHAOS_CASCADE_RATE"); cascade != "" {
		if r, err := strconv.ParseFloat(cascade, 64); err == nil {
			cfg.CascadeRate = r
		}
	}

	return cfg
}

// Stats tracks chaos injection statistics
type Stats struct {
	mu                sync.Mutex
	TotalChecks       int64
	InjectedFailures  int64
	FailuresByType    map[FailureType]int64
	FailuresByService map[string]int64
	CascadedFailures  int64
}

// Injector handles chaos injection
type Injector struct {
	cfg   *Config
	rng   *rand.Rand
	mu    sync.Mutex
	Stats Stats
}

// NewInjector creates a new chaos injector
func NewInjector(cfg *Config) *Injector {
	return &Injector{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		Stats: Stats{
			FailuresByType:    make(map[FailureType]int64),
			FailuresByService: make(map[string]int64),
		},
	}
}

// MaybeInject potentially injects a failure based on configuration
func (i *Injector) MaybeInject(ctx context.Context, service, operation string) error {
	return i.maybeInjectInternal(ctx, service, operation, false)
}

func (i *Injector) maybeInjectInternal(ctx context.Context, service, operation string, isCascade bool) error {
	i.Stats.mu.Lock()
	i.Stats.TotalChecks++
	i.Stats.mu.Unlock()

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
	i.mu.Lock()
	roll := i.rng.Float64()
	i.mu.Unlock()

	if roll > i.cfg.Rate {
		return nil
	}

	// Pick a failure type
	i.mu.Lock()
	failType := i.cfg.Types[i.rng.Intn(len(i.cfg.Types))]
	i.mu.Unlock()

	// Track stats
	i.Stats.mu.Lock()
	i.Stats.InjectedFailures++
	i.Stats.FailuresByType[failType]++
	i.Stats.FailuresByService[service]++
	if isCascade {
		i.Stats.CascadedFailures++
	}
	i.Stats.mu.Unlock()

	// Inject the failure
	err := i.injectFailure(ctx, service, operation, failType)

	// Maybe cascade to another failure
	if err != nil && i.cfg.CascadeRate > 0 {
		i.mu.Lock()
		cascadeRoll := i.rng.Float64()
		i.mu.Unlock()

		if cascadeRoll < i.cfg.CascadeRate {
			// Cascade: inject another failure
			cascadeErr := i.maybeInjectInternal(ctx, service, operation+"-cascade", true)
			if cascadeErr != nil {
				return fmt.Errorf("%w (cascaded: %v)", err, cascadeErr)
			}
		}
	}

	return err
}

func (i *Injector) injectFailure(ctx context.Context, service, operation string, failType FailureType) error {
	switch failType {
	case FailureTypeTimeout:
		// Simulate timeout by sleeping longer than typical timeouts
		select {
		case <-time.After(30 * time.Second):
			return errors.New("operation timed out")
		case <-ctx.Done():
			return ctx.Err()
		}

	case FailureTypeLatency:
		// Inject latency but don't fail
		i.mu.Lock()
		latencyMs := i.cfg.MinLatencyMs + i.rng.Intn(i.cfg.MaxLatencyMs-i.cfg.MinLatencyMs+1)
		i.mu.Unlock()

		select {
		case <-time.After(time.Duration(latencyMs) * time.Millisecond):
			return nil // Latency only, no error
		case <-ctx.Done():
			return ctx.Err()
		}

	case FailureTypePanic:
		panic(fmt.Sprintf("chaos: simulated panic in %s.%s", service, operation))

	case FailureTypePayment, FailureTypeFraud, FailureTypeInventory, FailureTypeShipping, FailureTypeValidation:
		// Use realistic error messages
		errors := RealisticErrors[failType]
		if len(errors) > 0 {
			i.mu.Lock()
			errMsg := errors[i.rng.Intn(len(errors))]
			i.mu.Unlock()
			return fmt.Errorf("%s", errMsg)
		}
		return fmt.Errorf("%s failure in %s.%s", failType, service, operation)

	case FailureTypeError:
		fallthrough
	default:
		return fmt.Errorf("chaos: injected failure in %s.%s", service, operation)
	}

	return nil
}

// InjectByType forces a specific failure type (for deterministic testing)
func (i *Injector) InjectByType(ctx context.Context, service, operation string, failType FailureType) error {
	i.Stats.mu.Lock()
	i.Stats.InjectedFailures++
	i.Stats.FailuresByType[failType]++
	i.Stats.FailuresByService[service]++
	i.Stats.mu.Unlock()

	return i.injectFailure(ctx, service, operation, failType)
}

// GetStats returns current chaos statistics
func (i *Injector) GetStats() Stats {
	i.Stats.mu.Lock()
	defer i.Stats.mu.Unlock()

	// Copy stats
	stats := Stats{
		TotalChecks:       i.Stats.TotalChecks,
		InjectedFailures:  i.Stats.InjectedFailures,
		CascadedFailures:  i.Stats.CascadedFailures,
		FailuresByType:    make(map[FailureType]int64),
		FailuresByService: make(map[string]int64),
	}
	for k, v := range i.Stats.FailuresByType {
		stats.FailuresByType[k] = v
	}
	for k, v := range i.Stats.FailuresByService {
		stats.FailuresByService[k] = v
	}
	return stats
}

// ResetStats clears all statistics
func (i *Injector) ResetStats() {
	i.Stats.mu.Lock()
	defer i.Stats.mu.Unlock()
	i.Stats.TotalChecks = 0
	i.Stats.InjectedFailures = 0
	i.Stats.CascadedFailures = 0
	i.Stats.FailuresByType = make(map[FailureType]int64)
	i.Stats.FailuresByService = make(map[string]int64)
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

// InitProduction initializes with production simulation settings
func InitProduction() {
	global = NewInjector(ProductionChaosConfig())
}

// MaybeInject uses the global injector
func MaybeInject(ctx context.Context, service, operation string) error {
	return global.MaybeInject(ctx, service, operation)
}

// InjectByType uses the global injector to force a specific failure
func InjectByTypeGlobal(ctx context.Context, service, operation string, failType FailureType) error {
	return global.InjectByType(ctx, service, operation, failType)
}

// Enabled checks if global chaos is enabled
func Enabled() bool {
	return global.Enabled()
}

// GetGlobalStats returns global injector statistics
func GetGlobalStats() Stats {
	return global.GetStats()
}

// ResetGlobalStats clears global statistics
func ResetGlobalStats() {
	global.ResetStats()
}

package main

import (
	"testing"

	"golang.org/x/mod/module"
)

func TestIsPseudoVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version string
		want    bool
	}{
		{"v1.30.0-148.4", false},
		{"v1.29.2", false},
		{"v1.30.0-rc.1", false},
		{"v0.0.0-20240101120000-abcdef123456", true},
		{"v1.29.1-0.20240101120000-abcdef123456", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()
			if got := module.IsPseudoVersion(tt.version); got != tt.want {
				t.Errorf("IsPseudoVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestValidateVersionConstraint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		serverVersion string
		latestRelease string
		wantErr       bool
	}{
		{"same minor", "v1.29.0-142.0", "v1.29.2", false},
		{"one minor ahead", "v1.30.0-148.4", "v1.29.2", false},
		{"two minors ahead", "v1.31.0-150.0", "v1.29.2", true},
		{"major ahead", "v2.0.0-1.0", "v1.29.2", true},
		{"exact match", "v1.29.2", "v1.29.2", false},
		{"invalid server", "invalid", "v1.29.2", true},
		{"invalid release", "v1.30.0", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateVersionConstraint(tt.serverVersion, tt.latestRelease)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVersionConstraint(%q, %q) error = %v, wantErr %v",
					tt.serverVersion, tt.latestRelease, err, tt.wantErr)
			}
		})
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

const (
	serverModule = "go.temporal.io/server"
	releaseURL   = "https://api.github.com/repos/temporalio/temporal/releases/latest"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	serverVersion, err := getServerVersion()
	if err != nil {
		return err
	}
	fmt.Printf("Found server dependency: %s@%s\n", serverModule, serverVersion)

	if module.IsPseudoVersion(serverVersion) {
		return fmt.Errorf("server dependency must be a tagged version, not a pseudo-version: %s", serverVersion)
	}
	fmt.Println("✓ Version is a valid tagged version")

	latestRelease, err := fetchLatestRelease()
	if err != nil {
		return err
	}
	fmt.Printf("Latest GitHub release: %s\n", latestRelease)

	if err := validateVersionConstraint(serverVersion, latestRelease); err != nil {
		return err
	}

	fmt.Println("✓ Server dependency version validation passed!")
	return nil
}

func getServerVersion() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to parse go.mod: %w", err)
	}

	for _, req := range f.Require {
		if req.Mod.Path == serverModule {
			return req.Mod.Version, nil
		}
	}

	return "", fmt.Errorf("server dependency %s not found in go.mod", serverModule)
}

func fetchLatestRelease() (string, error) {
	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// validateVersionConstraint ensures serverVersion is at most one minor version ahead of latestRelease.
func validateVersionConstraint(serverVersion, latestRelease string) error {
	if !semver.IsValid(serverVersion) {
		return fmt.Errorf("invalid server version: %s", serverVersion)
	}
	if !semver.IsValid(latestRelease) {
		return fmt.Errorf("invalid latest release version: %s", latestRelease)
	}

	serverMM := semver.MajorMinor(serverVersion)
	latestMM := semver.MajorMinor(latestRelease)

	var latestMajor, latestMinor int
	fmt.Sscanf(latestMM, "v%d.%d", &latestMajor, &latestMinor)
	maxAllowedMM := fmt.Sprintf("v%d.%d", latestMajor, latestMinor+1)

	fmt.Printf("  Server version: %s.x\n", serverMM)
	fmt.Printf("  Latest release: %s.x\n", latestMM)
	fmt.Printf("  Max allowed: %s.x\n", maxAllowedMM)

	if semver.Compare(serverMM, maxAllowedMM) > 0 {
		return fmt.Errorf(
			"server dependency version %s exceeds allowed range\n"+
				"  Max allowed: %s.x (latest release + 1 minor)\n"+
				"  Latest release: %s",
			serverVersion, maxAllowedMM, latestRelease,
		)
	}

	return nil
}

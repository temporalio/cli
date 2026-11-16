package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

const (
	serverModule = "go.temporal.io/server"
	apiModule    = "go.temporal.io/api"
	sdkModule    = "go.temporal.io/sdk"
	releaseURL   = "https://api.github.com/repos/temporalio/temporal/releases/latest"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	requires, err := readRequires("go.mod")
	if err != nil {
		return err
	}

	serverVersion, ok := requires[serverModule]
	if !ok {
		return fmt.Errorf("server dependency %s not found in go.mod", serverModule)
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

	if err := validateSDKVersions(requires); err != nil {
		return err
	}
	fmt.Println("✓ SDK dependencies are tagged versions")

	serverAPIVersion, err := getServerAPIVersion(serverVersion)
	if err != nil {
		return err
	}
	fmt.Printf("API version required by %s@%s: %s\n", serverModule, serverVersion, serverAPIVersion)

	if err := validateAPIVersion(requires[apiModule], serverAPIVersion); err != nil {
		return err
	}
	fmt.Println("✓ API dependency matches server")

	fmt.Println("✓ Server dependency version validation passed!")
	return nil
}

// readRequires returns the module path to version mapping of all requirements in a go.mod file.
func readRequires(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	f, err := modfile.Parse(path, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	requires := make(map[string]string, len(f.Require))
	for _, req := range f.Require {
		requires[req.Mod.Path] = req.Mod.Version
	}
	return requires, nil
}

// validateSDKVersions ensures the SDK and its contrib modules are pinned to tagged versions.
func validateSDKVersions(requires map[string]string) error {
	var paths []string
	for path := range requires {
		if path == sdkModule || strings.HasPrefix(path, sdkModule+"/") {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)

	for _, path := range paths {
		version := requires[path]
		if module.IsPseudoVersion(version) {
			return fmt.Errorf("SDK dependency must be a tagged version, not a pseudo-version: %s@%s", path, version)
		}
		fmt.Printf("  %s@%s\n", path, version)
	}
	return nil
}

// getServerAPIVersion returns the API version required by the given server version's go.mod.
func getServerAPIVersion(serverVersion string) (string, error) {
	// go mod download exits nonzero on failure but reports the cause in the JSON Error field,
	// so parse the output before checking the exit error.
	out, runErr := exec.Command("go", "mod", "download", "-json", serverModule+"@"+serverVersion).Output()

	var info struct {
		GoMod string
		Error string
	}
	parseErr := json.Unmarshal(out, &info)
	if parseErr == nil && info.Error != "" {
		return "", fmt.Errorf("failed to download %s@%s: %s", serverModule, serverVersion, info.Error)
	}
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("failed to download %s@%s: %w: %s",
				serverModule, serverVersion, runErr, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("failed to download %s@%s: %w", serverModule, serverVersion, runErr)
	}
	if parseErr != nil {
		return "", fmt.Errorf("failed to parse go mod download output: %w", parseErr)
	}

	requires, err := readRequires(info.GoMod)
	if err != nil {
		return "", err
	}
	version, ok := requires[apiModule]
	if !ok {
		return "", fmt.Errorf("%s not found in %s@%s go.mod", apiModule, serverModule, serverVersion)
	}
	return version, nil
}

// validateAPIVersion ensures the CLI uses exactly the API version the server was released with.
func validateAPIVersion(cliAPIVersion, serverAPIVersion string) error {
	if cliAPIVersion == "" {
		return fmt.Errorf("API dependency %s not found in go.mod", apiModule)
	}
	if cliAPIVersion != serverAPIVersion {
		return fmt.Errorf(
			"API dependency version %s does not match server\n"+
				"  %s: %s\n"+
				"  Required by %s: %s\n"+
				"  An SDK or other dependency bump may have raised the API version",
			cliAPIVersion, apiModule, cliAPIVersion, serverModule, serverAPIVersion,
		)
	}
	return nil
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

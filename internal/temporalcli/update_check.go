package temporalcli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"go.temporal.io/sdk/contrib/envconfig"
	"golang.org/x/mod/semver"
)

const updateCheckConfigProp = "cli.update_check.enabled"

var releaseVersionPattern = regexp.MustCompile(`/releases/download/(v[^/]+)/`)

type updateCheckState struct {
	Enabled        bool      `toml:"enabled,omitempty"`
	LatestVersion  string    `toml:"latest_version,omitempty"`
	LastCheckedAt  time.Time `toml:"last_checked_at,omitempty"`
	LastNotifiedAt time.Time `toml:"last_notified_at,omitempty"`
	NoticeCount    int       `toml:"notice_count,omitempty"`
}

type updateCheckHTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func resolveConfigFile(options CommandOptions) (path string, disabled bool) {
	for i, arg := range options.Args {
		switch {
		case arg == "--disable-config-file":
			disabled = true
		case arg == "--config-file" && i+1 < len(options.Args):
			path = options.Args[i+1]
		case strings.HasPrefix(arg, "--config-file="):
			path = strings.TrimPrefix(arg, "--config-file=")
		}
	}
	if path == "" && options.EnvLookup != nil {
		path, _ = options.EnvLookup.LookupEnv("TEMPORAL_CONFIG_FILE")
	}
	if path == "" {
		path = envconfig.DefaultConfigFilePath()
	}
	return path, disabled
}

func loadUpdateCheckState(path string) (map[string]any, updateCheckState, error) {
	raw := make(map[string]any)
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return raw, updateCheckState{}, nil
	}
	if err != nil {
		return nil, updateCheckState{}, err
	}
	if _, err := toml.Decode(string(b), &raw); err != nil {
		return nil, updateCheckState{}, err
	}
	var decoded struct {
		CLI struct {
			UpdateCheck updateCheckState `toml:"update_check"`
		} `toml:"cli"`
	}
	if _, err := toml.Decode(string(b), &decoded); err != nil {
		return nil, updateCheckState{}, err
	}
	return raw, decoded.CLI.UpdateCheck, nil
}

func storeUpdateCheckState(path string, raw map[string]any, state updateCheckState) error {
	cli, _ := raw["cli"].(map[string]any)
	if cli == nil {
		cli = make(map[string]any)
		raw["cli"] = cli
	}
	cli["update_check"] = state

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(raw); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".temporal.toml-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err := f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func setUpdateCheckEnabled(path string, enabled bool) error {
	raw, state, err := loadUpdateCheckState(path)
	if err != nil {
		return err
	}
	state.Enabled = enabled
	return storeUpdateCheckState(path, raw, state)
}

func updateCheckInterval(path string, checkedAt time.Time) time.Duration {
	h := fnv.New64a()
	_, _ = io.WriteString(h, path)
	_, _ = io.WriteString(h, checkedAt.UTC().Format("2006-01-02"))
	const jitterWindow = 24 * time.Hour
	jitter := time.Duration(h.Sum64()%uint64(jitterWindow)) - 12*time.Hour
	return 72*time.Hour + jitter
}

func noticeInterval(count int) time.Duration {
	switch count {
	case 1:
		return 24 * time.Hour
	case 2:
		return 72 * time.Hour
	case 3:
		return 7 * 24 * time.Hour
	default:
		return 14 * 24 * time.Hour
	}
}

func normalizedVersion(version string) string {
	version = strings.TrimSpace(version)
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if !semver.IsValid(version) {
		return ""
	}
	return version
}

func fetchLatestVersion(ctx context.Context, client updateCheckHTTPClient) (string, error) {
	url := fmt.Sprintf("https://temporal.download/cli/latest?platform=%s&arch=%s", runtime.GOOS, runtime.GOARCH)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "temporal-cli/"+Version)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %s", resp.Status)
	}
	var info struct {
		ArchiveURL string `json:"archiveUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	match := releaseVersionPattern.FindStringSubmatch(info.ArchiveURL)
	if len(match) != 2 || normalizedVersion(match[1]) == "" {
		return "", fmt.Errorf("download response contains no valid release version")
	}
	return match[1], nil
}

func runUpdateCheck(cctx *CommandContext) {
	current := normalizedVersion(Version)
	if current == "" || strings.Contains(strings.ToUpper(Version), "DEV") {
		return
	}
	path, disabled := resolveConfigFile(cctx.Options)
	if disabled {
		return
	}
	_, initialState, err := loadUpdateCheckState(path)
	if err != nil || !initialState.Enabled {
		return
	}

	lockPath := path + ".update-check.lock"
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > 5*time.Minute {
			_ = os.Remove(lockPath)
			lock, err = os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		}
	}
	if err != nil {
		return
	}
	lock.Close()
	defer os.Remove(lockPath)

	raw, state, err := loadUpdateCheckState(path)
	if err != nil || !state.Enabled {
		return
	}
	now := time.Now().UTC()
	if state.LastCheckedAt.IsZero() || now.Sub(state.LastCheckedAt) >= updateCheckInterval(path, state.LastCheckedAt) {
		// Persist the attempt before performing I/O so concurrent invocations do not
		// fan out when the endpoint is slow or unavailable.
		state.LastCheckedAt = now
		if err := storeUpdateCheckState(path, raw, state); err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(cctx, time.Second)
		latest, err := fetchLatestVersion(ctx, http.DefaultClient)
		cancel()
		if err == nil {
			latest = normalizedVersion(latest)
			if latest != state.LatestVersion {
				state.LatestVersion = latest
				state.LastNotifiedAt = time.Time{}
				state.NoticeCount = 0
			}
			_ = storeUpdateCheckState(path, raw, state)
		}
	}

	latest := normalizedVersion(state.LatestVersion)
	if latest == "" || semver.Compare(latest, current) <= 0 {
		return
	}
	if state.NoticeCount > 0 && now.Sub(state.LastNotifiedAt) < noticeInterval(state.NoticeCount) {
		return
	}
	fmt.Fprintf(cctx.Options.Stderr, "[notice] A new Temporal CLI release is available: %s -> %s\n", strings.TrimPrefix(current, "v"), strings.TrimPrefix(latest, "v"))
	fmt.Fprintf(cctx.Options.Stderr, "[notice] Release notes: https://github.com/temporalio/cli/releases/tag/%s\n", latest)
	state.LastNotifiedAt = now
	state.NoticeCount++
	_ = storeUpdateCheckState(path, raw, state)
}

package temporalcli

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type updateCheckClientFunc func(*http.Request) (*http.Response, error)

func (f updateCheckClientFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestUpdateCheckStateRoundTripPreservesConfig(t *testing.T) {
	path := t.TempDir() + "/temporal.toml"
	raw := map[string]any{
		"profile": map[string]any{"default": map[string]any{"address": "localhost:7233"}},
	}
	state := updateCheckState{
		Enabled:        true,
		LatestVersion:  "v1.8.0",
		LastCheckedAt:  time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC),
		LastNotifiedAt: time.Date(2026, 7, 11, 13, 0, 0, 0, time.UTC),
		NoticeCount:    2,
	}
	require.NoError(t, storeUpdateCheckState(path, raw, state))

	loadedRaw, loaded, err := loadUpdateCheckState(path)
	require.NoError(t, err)
	require.Equal(t, state, loaded)
	profiles := loadedRaw["profile"].(map[string]any)
	require.Equal(t, "localhost:7233", profiles["default"].(map[string]any)["address"])
}

func TestUpdateCheckIntervals(t *testing.T) {
	checkedAt := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	interval := updateCheckInterval("/tmp/temporal.toml", checkedAt)
	require.GreaterOrEqual(t, interval, 60*time.Hour)
	require.Less(t, interval, 84*time.Hour)
	require.Equal(t, 24*time.Hour, noticeInterval(1))
	require.Equal(t, 72*time.Hour, noticeInterval(2))
	require.Equal(t, 7*24*time.Hour, noticeInterval(3))
	require.Equal(t, 14*24*time.Hour, noticeInterval(4))
}

func TestFetchLatestVersion(t *testing.T) {
	client := updateCheckClientFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "temporal.download", req.URL.Host)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"archiveUrl":"https://temporal.download/assets/temporalio/cli/releases/download/v1.8.0/temporal_cli_1.8.0_linux_amd64.tar.gz"
			}`)),
		}, nil
	})
	version, err := fetchLatestVersion(context.Background(), client)
	require.NoError(t, err)
	require.Equal(t, "v1.8.0", version)
}

func TestNormalizedVersion(t *testing.T) {
	require.Equal(t, "v1.8.0", normalizedVersion("1.8.0"))
	require.Equal(t, "v1.8.0", normalizedVersion("v1.8.0"))
	require.Empty(t, normalizedVersion("not-a-version"))
}

func TestRunUpdateCheckUsesCachedVersionAndBacksOffNotice(t *testing.T) {
	path := t.TempDir() + "/temporal.toml"
	now := time.Now().UTC()
	require.NoError(t, storeUpdateCheckState(path, map[string]any{}, updateCheckState{
		Enabled:       true,
		LatestVersion: "v1.8.0",
		LastCheckedAt: now,
	}))

	originalVersion := Version
	Version = "1.7.0"
	t.Cleanup(func() { Version = originalVersion })
	var stderr bytes.Buffer
	cctx := &CommandContext{
		Context: context.Background(),
		Options: CommandOptions{
			Args:      []string{"--version", "--config-file", path},
			IOStreams: IOStreams{Stderr: &stderr},
		},
	}
	runUpdateCheck(cctx)
	require.Contains(t, stderr.String(), "1.7.0 -> 1.8.0")

	stderr.Reset()
	runUpdateCheck(cctx)
	require.Empty(t, stderr.String())
	_, state, err := loadUpdateCheckState(path)
	require.NoError(t, err)
	require.Equal(t, 1, state.NoticeCount)
}

package temporalcli

import (
	"os"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// Pinning modernc.org/sqlite to this version until https://gitlab.com/cznic/sqlite/-/issues/196 is resolved
func TestSqliteVersion(t *testing.T) {
	content, err := os.ReadFile("../go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "modernc.org/sqlite v1.34.1") {
		t.Errorf("go.mod missing dependency modernc.org/sqlite v1.34.1")
	}
	if !strings.Contains(contentStr, "modernc.org/libc v1.55.3") {
		t.Errorf("go.mod missing dependency modernc.org/libc v1.55.3")
	}
}

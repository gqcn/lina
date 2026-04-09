package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

// TestMain keeps package-level tests self-contained by generating the bundled
// runtime sample artifact before any test scans the shared plugin workspace.
func TestMain(m *testing.M) {
	if err := ensureBundledRuntimeSampleArtifactForTests(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to prepare bundled runtime sample: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func ensureBundledRuntimeSampleArtifactForTests() error {
	repoRoot, err := findRepoRoot(".")
	if err != nil {
		return err
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "plugin-demo-runtime")
	if _, statErr := os.Stat(filepath.Join(pluginDir, "plugin.yaml")); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil
		}
		return statErr
	}

	_, err = New().WriteRuntimeWasmArtifactFromSource(pluginDir)
	return err
}

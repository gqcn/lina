package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

// TestMain keeps package-level tests self-contained by generating the bundled
// dynamic sample artifact before any test scans the shared plugin workspace.
func TestMain(m *testing.M) {
	if err := ensureBundledRuntimeSampleArtifactForTests(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to prepare bundled dynamic sample: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func ensureBundledRuntimeSampleArtifactForTests() error {
	repoRoot, err := findRepoRoot(".")
	if err != nil {
		return err
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "plugin-demo-dynamic")
	if _, statErr := os.Stat(filepath.Join(pluginDir, "plugin.yaml")); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil
		}
		return statErr
	}

	builderDir := filepath.Join(repoRoot, "hack", "build-wasm")
	cmd := exec.Command(
		"go",
		"run",
		".",
		"--plugin-dir",
		pluginDir,
		"--output-dir",
		filepath.Join(repoRoot, "temp", "output"),
	)
	cmd.Dir = builderDir
	cmd.Env = append(os.Environ(), "GOWORK="+filepath.Join(repoRoot, "go.work"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run hack/build-wasm failed: %w: %s", err, string(output))
	}
	return nil
}

type testTopology struct {
	enabled bool
	primary bool
	nodeID  string
}

func (t *testTopology) IsEnabled() bool {
	return t != nil && t.enabled
}

func (t *testTopology) IsPrimary() bool {
	if t == nil {
		return true
	}
	return t.primary
}

func (t *testTopology) NodeID() string {
	if t == nil || t.nodeID == "" {
		return "test-node"
	}
	return t.nodeID
}

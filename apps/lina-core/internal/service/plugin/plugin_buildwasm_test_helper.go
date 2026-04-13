package plugin

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type runtimeBuildOutput struct {
	ArtifactPath string
	Content      []byte
}

func buildRuntimeArtifactWithHackTool(t *testing.T, pluginDir string) *runtimeBuildOutput {
	t.Helper()

	repoRoot, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	builderDir := filepath.Join(repoRoot, "hack", "build-wasm")
	cmd := exec.Command("go", "run", ".", "--plugin-dir", pluginDir)
	cmd.Dir = builderDir
	cmd.Env = append(os.Environ(), "GOWORK="+filepath.Join(repoRoot, "go.work"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run hack/build-wasm: %v output=%s", err, string(output))
	}

	manifest := &pluginManifest{}
	if err := New().loadPluginYAMLFile(filepath.Join(pluginDir, "plugin.yaml"), manifest); err != nil {
		t.Fatalf("failed to load plugin manifest after build: %v", err)
	}
	artifactPath := filepath.Join(pluginDir, "temp", buildPluginDynamicArtifactFileName(manifest.ID))
	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("failed to read built artifact: %v", err)
	}
	return &runtimeBuildOutput{
		ArtifactPath: artifactPath,
		Content:      content,
	}
}

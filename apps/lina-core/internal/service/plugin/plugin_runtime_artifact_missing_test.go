package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanPluginManifestsDiscoversRuntimePluginFromStorage(t *testing.T) {
	service := New()

	pluginID := "plugin-runtime-storage-scan"
	createTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Storage Scan Plugin",
		"v0.9.1",
		nil,
		nil,
	)

	manifests, err := service.scanPluginManifests()
	if err != nil {
		t.Fatalf("expected scan to discover runtime artifact from storage path, got error: %v", err)
	}

	for _, manifest := range manifests {
		if manifest == nil || manifest.ID != pluginID {
			continue
		}
		if manifest.RuntimeArtifact == nil {
			t.Fatalf("expected runtime artifact metadata to be loaded for %s", pluginID)
		}
		return
	}
	t.Fatalf("expected runtime plugin %s to be discovered from storage path", pluginID)
}

func TestScanPluginManifestsDropsRuntimePluginAfterArtifactRemoval(t *testing.T) {
	service := New()

	pluginID := "plugin-runtime-missing-scan"
	artifactPath := createTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Missing Scan Plugin",
		"v0.9.2",
		nil,
		nil,
	)

	if err := os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove generated runtime artifact: %v", err)
	}

	manifests, err := service.scanPluginManifests()
	if err != nil {
		t.Fatalf("expected scan to succeed after runtime artifact removal, got error: %v", err)
	}

	for _, manifest := range manifests {
		if manifest != nil && manifest.ID == pluginID {
			t.Fatalf("expected removed runtime plugin %s to disappear from scan results", pluginID)
		}
	}
}

func TestEnsureRuntimePluginArtifactAvailableRejectsMissingGeneratedWasm(t *testing.T) {
	service := New()

	pluginID := "plugin-runtime-missing-install"
	artifactPath := createTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Missing Install Plugin",
		"v0.9.3",
		nil,
		nil,
	)

	if err := os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove generated runtime artifact: %v", err)
	}

	manifest := &pluginManifest{
		ID:           pluginID,
		Name:         "Runtime Missing Install Plugin",
		Version:      "v0.9.3",
		Type:         pluginTypeRuntime.String(),
		ManifestPath: "",
		RootDir:      filepath.Dir(artifactPath),
	}

	strictErr := service.validateRuntimePluginArtifact(manifest, filepath.Dir(artifactPath))
	if strictErr == nil || !isMissingRuntimePluginArtifactError(strictErr) {
		t.Fatalf("expected strict runtime validation to report a missing artifact, got: %v", strictErr)
	}

	err := service.ensureRuntimePluginArtifactAvailable(manifest, "安装")
	if err == nil {
		t.Fatalf("expected lifecycle guard to reject missing runtime artifact")
	}
	if expected := "make wasm p=" + pluginID; !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected lifecycle guard error to mention %q, got: %v", expected, err)
	}
	if expected := filepath.ToSlash(buildPluginRuntimeArtifactRelativePath(pluginID)); !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected lifecycle guard error to mention missing wasm path %q, got: %v", expected, err)
	}
}

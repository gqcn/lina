package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"lina-build-wasm/builder"
)

func main() {
	var pluginDir string
	flag.StringVar(&pluginDir, "plugin-dir", "", "Runtime plugin directory")
	flag.Parse()

	if pluginDir == "" {
		fmt.Fprintln(os.Stderr, "missing --plugin-dir")
		os.Exit(1)
	}

	absolutePluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve plugin dir: %v\n", err)
		os.Exit(1)
	}

	out, err := builder.WriteRuntimeWasmArtifactFromSource(absolutePluginDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build runtime artifact: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("built runtime artifact: %s\n", out.ArtifactPath)
}

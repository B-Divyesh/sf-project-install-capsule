package capsule

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectedEngineMustReportRootless(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-podman")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf '%s\\n' \"$ROOTLESS_REPLY\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CAPSULE_ENGINE", path)
	t.Setenv("ROOTLESS_REPLY", "true")
	engine, err := FindEngine(context.Background(), true)
	if err != nil || engine.Kind != "podman" {
		t.Fatalf("rootless engine rejected: %#v, %v", engine, err)
	}
	t.Setenv("ROOTLESS_REPLY", "false")
	_, err = FindEngine(context.Background(), true)
	if err == nil || !strings.Contains(err.Error(), "not rootless") {
		t.Fatalf("unsafe engine accepted: %v", err)
	}
}

package cmd

import (
	"path/filepath"
	"testing"
)

func TestArtifactOutputPathPreservesRelativeDirectories(t *testing.T) {
	root := t.TempDir()
	frontend, err := artifactOutputPath(root, "frontend/report.xml", "report.xml")
	if err != nil {
		t.Fatal(err)
	}
	backend, err := artifactOutputPath(root, "backend/report.xml", "report.xml")
	if err != nil {
		t.Fatal(err)
	}
	if frontend == backend {
		t.Fatal("artifacts with the same filename should not overwrite each other")
	}
	if filepath.Dir(frontend) == root {
		t.Fatalf("relative artifact directory was discarded: %s", frontend)
	}
}

func TestArtifactOutputPathRejectsTraversal(t *testing.T) {
	if _, err := artifactOutputPath(t.TempDir(), "../secret.txt", "secret.txt"); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
}

package cmd

import (
	"os"
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

func TestPrepareArtifactPathRejectsParentSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "escaped")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	out := filepath.Join(link, "artifact.zip")
	if err := prepareArtifactPath(root, out); err == nil {
		t.Fatal("expected parent symlink escape to be rejected")
	}
}

func TestPrepareArtifactPathRejectsDestinationSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "artifact.zip")
	if err := os.Symlink(target, out); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := prepareArtifactPath(root, out); err == nil {
		t.Fatal("expected destination symlink to be rejected")
	}
}

func TestArtifactOutputPathRejectsTraversal(t *testing.T) {
	if _, err := artifactOutputPath(t.TempDir(), "../secret.txt", "secret.txt"); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
}

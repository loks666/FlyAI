package webui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromConfiguredDist(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "_next"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WEB_DIST_PATH", root)

	assets, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	content, err := assets.Read("index.html")
	if err != nil || string(content) != "ok" {
		t.Fatalf("Read(index.html) = %q, %v", content, err)
	}
}

func TestReadRejectsParentTraversal(t *testing.T) {
	assets := &Assets{root: t.TempDir()}
	if _, err := assets.Read("../secret"); err == nil {
		t.Fatal("expected parent traversal to be rejected")
	}
}

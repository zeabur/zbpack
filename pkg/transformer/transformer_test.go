package transformer_test

import (
	"os"
	"path/filepath"
	"testing"
)

// GetOutputSnapshotPath returns the path to the output snapshot.
func GetOutputSnapshotPath(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	snapshotDir := filepath.Join(wd, "snapshot", t.Name())
	err = os.MkdirAll(snapshotDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	return snapshotDir
}

// GetInputPath returns the path to the input.
func GetInputPath(t *testing.T, name string) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	return filepath.Join(wd, "inputs", name)
}

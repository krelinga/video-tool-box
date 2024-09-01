package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func setUpTempDir(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "state_test.go")
	if err != nil {
		t.Fatal("could not create tempdir")
	}
	return tempDir
}

func tearDownTempDir(t *testing.T, tempDir string) {
	t.Helper()
	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("could not delete tempDir: %s", err)
	}
}

func TestReadNonExistingStateFile(t *testing.T) {
	t.Parallel()
	tempDir := setUpTempDir(t)
	defer tearDownTempDir(t, tempDir)

	tsPath := filepath.Join(tempDir, "does_not_exist")
	ts, err := readToolState(tsPath)
	if err != nil {
		t.Error(err)
	}
	if len(ts.Projects) != 0 {
		t.Error(ts.Projects)
	}
}

func TestCorruptStateFile(t *testing.T) {
	t.Parallel()
	tempDir := setUpTempDir(t)
	defer tearDownTempDir(t, tempDir)

	tsPath := filepath.Join(tempDir, "corrupt")
	corrupt := []byte("THIS IS NOT JSON")
	if err := os.WriteFile(tsPath, corrupt, 0644); err != nil {
		t.Fatal(err)
	}
	_, err := readToolState(tsPath)
	if err == nil {
		t.Error("expected error")
	}
}

func TestStateFileWrites(t *testing.T) {
	t.Parallel()
	tempDir := setUpTempDir(t)
	defer tearDownTempDir(t, tempDir)

	tsPath := filepath.Join(tempDir, "state")
	ts1 := &toolState{
		Projects: []*projectState{
			{
				Pt:   ptMovie,
				Name: "movie",
			},
		},
	}
	if err := writeToolState(ts1, tsPath); err != nil {
		t.Errorf("error writing to non-existing state file: %s", err)
	}

	ts1Read, err := readToolState(tsPath)
	if err != nil {
		t.Errorf("error reading toolState: %s", err)
	}
	if !cmp.Equal(ts1, ts1Read) {
		t.Error(cmp.Diff(ts1, ts1Read))
	}

	ts2 := &toolState{
		Projects: []*projectState{
			{
				Pt:   ptShow,
				Name: "show",
			},
		},
	}
	if err := writeToolState(ts2, tsPath); err != nil {
		t.Errorf("error overwriting existing state file: %s", err)
	}
	ts2Read, err := readToolState(tsPath)
	if err != nil {
		t.Errorf("error reading toolState: %s", err)
	}
	if !cmp.Equal(ts2, ts2Read) {
		t.Error(cmp.Diff(ts2, ts2Read))
	}
}

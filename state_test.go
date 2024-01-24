package main

import (
    "os"
    "path/filepath"
    "testing"
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
    ts, err := newToolState(tsPath)
    if err != nil {
        t.Error(err)
    }
    if ts.Pt != ptUndef {
        t.Error(ts.Pt)
    }
    if ts.Name != "" {
        t.Error(ts.Name)
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
    _, err := newToolState(tsPath)
    if err == nil {
        t.Error("expected error")
    }
}

func TestReadAndWriteStateFile(t *testing.T) {
    tempDir := setUpTempDir(t)
    defer tearDownTempDir(t, tempDir)
    if err := os.Setenv("HOME", tempDir) ; err != nil {
        t.Fatal("could not set fake homedir")
    }

    errorToolState := toolState {
        Name: "this should be removed.",
    }
    gToolState = errorToolState
    if err := readToolState(); err != nil {
        t.Fatalf("error reading non-existing state file: %s", err)
    }
    if gToolState != (toolState{}) {
        t.Error("existing gToolState not cleared when reading non-existing file.")
    }

    gToolState.Name = "initialWrite"
    toolStateThatWasWritten := gToolState
    if err := writeToolState(); err != nil {
        t.Fatalf("error writing new state file: %s", err)
    }

    if err := readToolState(); err != nil {
        t.Fatalf("could not read from existing state file: %s", err)
    }
    if gToolState != toolStateThatWasWritten {
        t.Fatalf("inconsistent state file, expected %s and saw %s", toolStateThatWasWritten, gToolState)
    }
}

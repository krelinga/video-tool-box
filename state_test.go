package main

import (
    "os"
    "testing"
)

var (
    tempDir string
)

func setUp(t *testing.T) {
    t.Helper()
    tempDir, err := os.MkdirTemp("", "state_test.go")
    if err != nil {
        t.Fatal("could not create fake homedir")
    }
    if err := os.Setenv("HOME", tempDir) ; err != nil {
        t.Fatal("could not set fake homedir")
    }
}

func tearDown(t *testing.T) {
    t.Helper()
    if err := os.RemoveAll(tempDir); err != nil {
        t.Fatalf("could not delete tempDir: %s", err)
    }
}

func TestReadAndWriteStateFile(t *testing.T) {
    setUp(t)
    defer tearDown(t)

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

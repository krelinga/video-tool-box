package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	// TODO: consolidate testing utils for managing temporary directories with
	// state_test.go
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	}()

	t.Run("non_existing_config", func(t *testing.T) {
		if _, err := readConfig("/does/not/exist"); err != nil {
			t.Errorf("Expected no error when reading non-existing file: %s", err)
		}
	})
	t.Run("round_trip", func(t *testing.T) {
		path := filepath.Join(tmpDir, "round_trip.json")
		c := &config{
			MkvUtilServerTarget: "/foo/bar",
		}
		if err := writeConfig(c, path); err != nil {
			t.Error(err)
			return
		}
		readC, err := readConfig(path)
		if err != nil {
			t.Error(err)
			return
		}
		if !cmp.Equal(c, readC) {
			t.Errorf(cmp.Diff(c, readC))
		}
	})
}

func TestConfigString(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		expected := `MkvUtilServerTarget: <empty>
TcServerTarget: <empty>
DefaultShowTranscodeOutDir: <empty>
DefaultMovieTranscodeOutDir: <empty>
RipCacheServerDir: <empty>
RipCacheLocalDir: <empty>
`
		c := &config{}
		actual := c.String()
		if !cmp.Equal(expected, actual) {
			t.Error(cmp.Diff(expected, actual))
		}
	})
	t.Run("set", func(t *testing.T) {
		t.Parallel()
		expected := `MkvUtilServerTarget: "taters"
TcServerTarget: <empty>
DefaultShowTranscodeOutDir: <empty>
DefaultMovieTranscodeOutDir: <empty>
RipCacheServerDir: <empty>
RipCacheLocalDir: <empty>
`
		c := &config{
			MkvUtilServerTarget: "taters",
		}
		actual := c.String()
		if !cmp.Equal(expected, actual) {
			t.Error(cmp.Diff(expected, actual))
		}
	})
}

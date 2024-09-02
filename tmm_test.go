package main

// spell-checker:ignore tvshow.nfo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/krelinga/video-tool-box/nfo"
)

func TestFindNfoFiles(t *testing.T) {
	t.Parallel()

	// Create temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"file1.nfo",
		"file2.txt",
		"file3.nfo",
		"tvshow.nfo",
		"file5.xml",
	}
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte(file), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Call the function being tested
	files, err := findNfoFiles(tempDir)
	if err != nil {
		t.Fatalf("findNfoFiles returned an error: %v", err)
	}

	// Check the number of files found
	expectedCount := 2
	if len(files) != expectedCount {
		t.Errorf("expected %d .nfo files, got %d", expectedCount, len(files))
	}

	// Check the paths of the found files
	expectedFiles := []string{
		"file1.nfo",
		"file3.nfo",
	}
	for i, file := range files {
		expectedPath := filepath.Join(tempDir, expectedFiles[i])
		if file.path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, file.path)
		}
		if file.content != expectedFiles[i] {
			t.Errorf("expected content %s, got %s", expectedFiles[i], file.content)
		}
	}
}
func TestGuessTranscodeProfile(t *testing.T) {
	testCases := []struct {
		input    *nfo.Content
		expected string
	}{
		{
			input: &nfo.Content{
				Width:  1920,
				Height: 1080,
			},
			expected: "hd",
		},
		{
			input: &nfo.Content{
				Width:  1280,
				Height: 720,
				Genres: []string{"anime"},
			},
			expected: "sd_animation",
		},
		{
			input: &nfo.Content{
				Width:  1280,
				Height: 720,
				Genres: []string{"animation"},
			},
			expected: "sd_animation",
		},
		{
			input: &nfo.Content{
				Width:  1280,
				Height: 720,
				Tags:   []string{"Anime"},
			},
			expected: "sd_animation",
		},
		{
			input: &nfo.Content{
				Width:  1280,
				Height: 720,
				Tags:   []string{"adult animation"},
			},
			expected: "sd_animation",
		},
		{
			input: &nfo.Content{
				Width:  1280,
				Height: 720,
				Genres: []string{"action"},
			},
			expected: "sd_live_action",
		},
		{
			input: &nfo.Content{
				Width:  1280,
				Height: 720,
				Tags:   []string{"action"},
			},
			expected: "sd_live_action",
		},
	}

	for _, tc := range testCases {
		actual := guessTranscodeProfile(tc.input)
		if actual != tc.expected {
			t.Errorf("expected profile %s, got %s", tc.expected, actual)
		}
	}
}
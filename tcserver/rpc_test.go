package main

import (
    "os"
    "path/filepath"
    "testing"

    //"github.com/krelinga/video-tool-box/pb"
)

func TestOutputPathExists(t *testing.T) {
    t.Parallel()
    tmpDir, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatal(err)
    }
    defer func() {
        err := os.RemoveAll(tmpDir)
        if err != nil {
            t.Fatal(err)
        }
    }()
    touch := func(path string) {
        if err := os.WriteFile(path, []byte(""), 0777); err != nil {
            t.Fatal(err)
        }
    }
    inPath := filepath.Join(tmpDir, "in.mkv")
    outPath := filepath.Join(tmpDir, "out.mkv")
    touch(inPath)
    touch(outPath)
}

package main

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/krelinga/video-tool-box/pb"
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

    statePath := filepath.Join(tmpDir, "state")
    defaultProfile := "mkv_h265_1080p30"
    tcServer := newTcServer(statePath, defaultProfile)


    startReq := &pb.StartAsyncTranscodeRequest{
        Name: "test",
        InPath: inPath,
        OutPath: outPath,
    }
    _, err = tcServer.StartAsyncTranscode(context.Background(), startReq)
    if err != nil {
        t.Fatal(err)
    }
}

package main

import (
    "context"
    "os"
    "path/filepath"
    "strings"
    "time"
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


    const name = "test"
    startReq := &pb.StartAsyncTranscodeRequest{
        Name: name,
        InPath: inPath,
        OutPath: outPath,
    }
    _, err = tcServer.StartAsyncTranscode(context.Background(), startReq)
    if err != nil {
        t.Fatal(err)
    }

    checkForError := func() (retry bool) {
        checkReq := &pb.CheckAsyncTranscodeRequest{
            Name: name,
        }
        checkReply, err := tcServer.CheckAsyncTranscode(context.Background(), checkReq)
        if err != nil {
            t.Fatal(err)
        }
        if checkReply.State == pb.CheckAsyncTranscodeReply_STATE_IN_PROGRESS {
            retry = true
            return
        }
        failed := checkReply.State == pb.CheckAsyncTranscodeReply_STATE_FAILED
        correctError := strings.Contains(checkReply.ErrorMessage, "already exists")
        if failed && correctError {
            // This is our expected case.
            retry = false
            return
        }
        // Otherwise we ended up in some unexpected case.
        t.Error(checkReply)
        retry = false
        return
    }
    for checkForError() {
        t.Log("transcode in-progress, retrying...")
        time.Sleep(time.Millisecond * 100)
    }
}

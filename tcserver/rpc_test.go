package main

import (
    "context"
    "os"
    "path/filepath"
    "strings"
    "time"
    "testing"

    "github.com/krelinga/video-tool-box/pb"
    "github.com/krelinga/video-tool-box/tcserver/transcoder"
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

    const defaultProfile = "mkv_h265_1080p30"
    tran := transcoder.Transcoder{
        FileWorkers: 1,
        MaxQueuedFiles: 10000,
        ShowWorkers: 1,
        MaxQueuedShows: 1,
    }
    if err := tran.Start(); err != nil {
        t.Fatal(err)
    }
    defer tran.Stop()
    tcServer := newTcServer(defaultProfile, &tran)

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
        isInProgress := checkReply.State == pb.TranscodeState_IN_PROGRESS
        isNotStarted := checkReply.State == pb.TranscodeState_NOT_STARTED
        if isInProgress || isNotStarted {
            retry = true
            return
        }
        failed := checkReply.State == pb.TranscodeState_FAILED
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

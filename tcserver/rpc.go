package main

import (
    "context"
    "errors"
    "fmt"
    "os"
    "os/exec"

    "github.com/krelinga/video-tool-box/pb"
)

type tcServer struct {
    pb.UnimplementedTCServerServer
}

func (tcs *tcServer) HelloWorld(ctx context.Context, req *pb.HelloWorldRequest) (*pb.HelloWorldReply, error) {
    fmt.Printf("Saw request: %v\n", req)
    fileSize := func() int64 {
        translated, err := translatePath(req.In)
        if err != nil {
            return -1
        }
        stat, err := os.Stat(translated)
        if err != nil {
            return -1
        }
        return stat.Size()
    }()
    rep := &pb.HelloWorldReply{
        Out: req.In,
        FileSize: fileSize,
    }
    return rep, nil
}

type handbrakeFlags []string

var gHandbrakeProfile = map[string]handbrakeFlags{
    "mkv_h265_1080p30": {
        "-Z", "Matroska/H.265 MKV 1080p30",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
}

func (tcs *tcServer) TranscodeOneFile(ctc context.Context, req *pb.TranscodeOneFileRequest) (*pb.TranscodeOneFileReply, error) {
    fmt.Printf("TranscodeOneFile: %v\n", req)
    inPath, err := translatePath(req.InPath)
    if err != nil {
        return nil, err
    }
    outPath, err := translatePath(req.OutPath)
    if err != nil {
        return nil, err
    }

    const profile = "mkv_h265_1080p30"
    profileFlags, ok := gHandbrakeProfile[profile]
    if !ok {
        return nil, errors.New(fmt.Sprintf("unknown profile %s", profile))
    }
    standardFlags := []string{
        "-i", inPath,
        "-o", outPath,
    }

    cmd := exec.Command("/usr/bin/HandBrakeCLI")
    cmd.Args = append(cmd.Args, standardFlags...)
    cmd.Args = append(cmd.Args, profileFlags...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return &pb.TranscodeOneFileReply{}, cmd.Run()
}

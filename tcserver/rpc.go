package main

import (
    "context"
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

    // Temporary, until we get handbrake hooked up.
    copyCmd := exec.Command("cp", inPath, outPath)
    return &pb.TranscodeOneFileReply{}, copyCmd.Run()
}

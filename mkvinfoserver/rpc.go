package main

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"

    "github.com/krelinga/video-tool-box/pb"
)

type MkvInfoServer struct {
    pb.UnimplementedMkvInfoServer
}

func (mis *MkvInfoServer) GetMkvInfo(ctx context.Context, req *pb.GetMkvInfoRequest) (*pb.GetMkvInfoReply, error) {
    fmt.Printf("Saw request: %v\n", req)
    cmd := exec.Command("mkvinfo", req.In)
    b := &bytes.Buffer{}
    cmd.Stdout = b
    cmd.Stderr = b
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("Could not run mkvinfo tool: %w", err)
    }
    reply := &pb.GetMkvInfoReply{
        Out: b.String(),
    }
    return reply, nil
}



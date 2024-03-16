package main

import (
    "context"
    "fmt"
    "os"

    "github.com/krelinga/video-tool-box/pb"
)

type MkvInfoServer struct {
    pb.UnimplementedMkvInfoServer
}

func (mis *MkvInfoServer) GetMkvInfo(ctx context.Context, req *pb.GetMkvInfoRequest) (*pb.GetMkvInfoReply, error) {
    fmt.Printf("Saw request: %v\n", req)
    if _, err := os.Stat(req.In); err != nil {
        return nil, fmt.Errorf("Could not stat %s: %w", req.In, err)
    }
    reply := &pb.GetMkvInfoReply{
        Out: "bar",
    }
    return reply, nil
}



package main

import (
    "context"
    "fmt"

    "github.com/krelinga/video-tool-box/pb"
)

type MkvInfoServer struct {
    pb.UnimplementedMkvInfoServer
}

func (mis *MkvInfoServer) GetMkvInfo(ctx context.Context, req *pb.GetMkvInfoRequest) (*pb.GetMkvInfoReply, error) {
    fmt.Printf("Saw request: %v\n", req)
    reply := &pb.GetMkvInfoReply{
        Out: "bar",
    }
    return reply, nil
}



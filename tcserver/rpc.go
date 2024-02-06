package main

import (
    "context"
    "fmt"
    "os"

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


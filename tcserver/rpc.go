package main

import (
    "context"
    "fmt"
    "os"

    "github.com/krelinga/video-tool-box/pb"
    "github.com/krelinga/video-tool-box/tcserver/transcoder"
)

type tcServer struct {
    pb.UnimplementedTCServerServer

    defaultProfile  string
    tc* transcoder.Transcoder
}

func newTcServer(defaultProfile string, tc *transcoder.Transcoder) *tcServer {
    return &tcServer{
        defaultProfile: defaultProfile,
        tc: tc,
    }
}

func (tcs *tcServer) HelloWorld(ctx context.Context, req *pb.HelloWorldRequest) (*pb.HelloWorldReply, error) {
    fmt.Printf("Saw request: %v\n", req)
    fileSize := func() int64 {
        stat, err := os.Stat(req.In)
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

func (tcs *tcServer) StartAsyncTranscode(ctx context.Context, req *pb.StartAsyncTranscodeRequest) (*pb.StartAsyncTranscodeReply, error) {
    fmt.Printf("StartAsyncTranscode: %v\n", req)
    profile := func() string {
        if len(req.Profile) > 0 {
            return req.Profile
        }
        return tcs.defaultProfile
    }()
    fmt.Printf("using profile %s\n", profile)
    err := tcs.tc.StartFile(req.Name, req.InPath, req.OutPath, profile)
    return &pb.StartAsyncTranscodeReply{}, err
}

func (tcs *tcServer) CheckAsyncTranscode(ctx context.Context, req *pb.CheckAsyncTranscodeRequest) (*pb.CheckAsyncTranscodeReply, error) {
    fmt.Printf("CheckAsyncTranscode: %v\n", req)
    reply := &pb.CheckAsyncTranscodeReply{}
    readState := func(s *transcoder.SingleFileState) {
        switch s.St {
        case transcoder.StateNotStarted:
            reply.State = pb.TranscodeState_NOT_STARTED
        case transcoder.StateInProgress:
            reply.State = pb.TranscodeState_IN_PROGRESS
            if s.Latest != nil {
                reply.Progress = s.Latest.String()
            }
        case transcoder.StateComplete:
            reply.State = pb.TranscodeState_DONE
        case transcoder.StateError:
            reply.State = pb.TranscodeState_FAILED
            reply.ErrorMessage = s.Err.Error()
        default:
            panic(s.St)
        }
    }
    return reply, tcs.tc.CheckFile(req.Name, readState)
}

func (tcs *tcServer) StartAsyncShowTranscode(ctx context.Context, req *pb.StartAsyncShowTranscodeRequest) (*pb.StartAsyncShowTranscodeReply, error) {
    fmt.Printf("StartAsyncShowTranscode: %v\n", req)
    profile := func() string {
        if len(req.Profile) > 0 {
            return req.Profile
        }
        return tcs.defaultProfile
    }()
    fmt.Printf("using profile %s\n", profile)

    return nil, nil
}

func (tcs *tcServer) CheckAsyncShowTranscode(ctx context.Context, req *pb.CheckAsyncShowTranscodeRequest) (*pb.CheckAsyncShowTranscodeReply, error) {
    return nil, nil
}

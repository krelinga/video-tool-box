package main

import (
    "context"
    "fmt"
    "errors"
    "os"
    "path/filepath"

    "github.com/krelinga/video-tool-box/pb"
    "github.com/krelinga/video-tool-box/tcserver/hb"
)

type tcServer struct {
    pb.UnimplementedTCServerServer

    defaultProfile  string
    s               *state
}

func newTcServer(profile string) *tcServer {
    return &tcServer{
        s: &state{},
        defaultProfile: profile,
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

func transcodeImpl(inPath, outPath, profile string, s *state) error {
    prog := func(u *hb.Progress) {
        s.Do(func(_ *pb.TCSState, prog **hb.Progress) error {
            *prog = u
            return nil
        })
    }
    return hb.Run(inPath, outPath, profile, prog)
}

// Starts Handbrake and blocks until it finishes.
func (tcs *tcServer) transcode(inPath, outPath, profile string) {
    err := func() error {
        if err := os.MkdirAll(filepath.Dir(outPath), 0777); err != nil {
            return err
        }
        if err := copyRelatedFiles(inPath, outPath); err != nil {
            return err
        }
        if err := transcodeImpl(inPath, outPath, profile, tcs.s); err != nil {
            return err
        }
        return nil
    }()
    persistErr := tcs.s.Do(func(sp *pb.TCSState, prog **hb.Progress) error {
        if err != nil {
            sp.Op.State = pb.TCSState_Op_STATE_FAILED
            sp.Op.ErrorMessage = err.Error()
        } else {
            sp.Op.State = pb.TCSState_Op_STATE_DONE
        }
        *prog = nil
        return nil
    })
    if persistErr != nil {
        // TODO: Is there a better way here?
        panic(persistErr.Error())
    }
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
    err := tcs.s.Do(func(sp *pb.TCSState, _ **hb.Progress) error {
        if sp.Op != nil && sp.Op.State == pb.TCSState_Op_STATE_IN_PROGRESS {
            return fmt.Errorf("Async transcode %s already in-progress", sp.Op.Name)
        }
        sp.Op = &pb.TCSState_Op{
            Name: req.Name,
            State: pb.TCSState_Op_STATE_IN_PROGRESS,
        }
        go tcs.transcode(req.InPath, req.OutPath, profile)
        return nil
    })
    return &pb.StartAsyncTranscodeReply{}, err
}

func (tcs *tcServer) CheckAsyncTranscode(ctx context.Context, req *pb.CheckAsyncTranscodeRequest) (*pb.CheckAsyncTranscodeReply, error) {
    fmt.Printf("CheckAsyncTranscode: %v\n", req)
    reply := &pb.CheckAsyncTranscodeReply{}
    err := tcs.s.Do(func(sp *pb.TCSState, prog **hb.Progress) error {
        if sp.Op == nil || sp.Op.State == pb.TCSState_Op_STATE_UNKNOWN {
            return errors.New("No active transcode")
        }
        if sp.Op.Name != req.Name {
            return fmt.Errorf("Active transcode is named '%s', but '%s' was requested", sp.Op.Name, req.Name)
        }
        reply.State = func() pb.TranscodeState {
            switch sp.Op.State {
            case pb.TCSState_Op_STATE_UNKNOWN:
                return pb.TranscodeState_UNKNOWN
            case pb.TCSState_Op_STATE_IN_PROGRESS:
                return pb.TranscodeState_IN_PROGRESS
            case pb.TCSState_Op_STATE_DONE:
                return pb.TranscodeState_DONE
            case pb.TCSState_Op_STATE_FAILED:
                return pb.TranscodeState_FAILED
            default:
                panic(fmt.Sprintf("Unexpected op state %v", sp.Op.State))
            }
        }()
        reply.ErrorMessage = sp.Op.ErrorMessage
        if *prog != nil {
            reply.Progress = (*prog).String()
        }
        return nil
    })
    return reply, err
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

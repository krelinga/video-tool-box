package main

import (
    "context"
    "fmt"
    "errors"
    "os"
    "os/exec"

    "github.com/krelinga/video-tool-box/pb"
)

type tcServer struct {
    pb.UnimplementedTCServerServer

    s   *state
}

func newTcServer(statePath string) *tcServer {
    return &tcServer{s: newState(statePath)}
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

func transcodeImpl(inNasPath, outNasPath, profile string) error {
    inPath, err := translatePath(inNasPath)
    if err != nil {
        return err
    }
    outPath, err := translatePath(outNasPath)
    if err != nil {
        return err
    }

    profileFlags, ok := gHandbrakeProfile[profile]
    if !ok {
        return fmt.Errorf("unknown profile %s", profile)
    }
    standardFlags := []string{
        "-i", inPath,
        "-o", outPath,
        "--json",
    }

    stdOutPath := outPath + ".stdout"
    stdOutFile, err := os.Create(stdOutPath)
    if err != nil {
        return fmt.Errorf("could not open %s: %v", stdOutPath, err)
    }
    defer stdOutFile.Close()

    stdErrPath := outPath + ".stderr"
    stdErrFile, err := os.Create(stdErrPath)
    if err != nil {
        return fmt.Errorf("could not open %s: %v", stdErrPath, err)
    }
    defer stdErrFile.Close()

    cmd := exec.Command("/usr/bin/HandBrakeCLI")
    cmd.Args = append(cmd.Args, standardFlags...)
    cmd.Args = append(cmd.Args, profileFlags...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = stdOutFile
    cmd.Stderr = stdErrFile
    return cmd.Run()
}

func (tcs *tcServer) StartAsyncTranscode(ctx context.Context, req *pb.StartAsyncTranscodeRequest) (*pb.StartAsyncTranscodeReply, error) {
    fmt.Printf("StartAsyncTranscode: %v\n", req)
    const profile = "mkv_h265_1080p30"
    err := tcs.s.Do(func(sp *pb.TCSState) error {
        if sp.Op != nil && sp.Op.State == pb.TCSState_Op_STATE_IN_PROGRESS {
            return fmt.Errorf("Async transcode %s already in-progress", sp.Op.Name)
        }
        sp.Op = &pb.TCSState_Op{
            Name: req.Name,
            State: pb.TCSState_Op_STATE_IN_PROGRESS,
        }
        go func() {
            err := transcodeImpl(req.InPath, req.OutPath, profile)
            persistErr := tcs.s.Do(func(sp *pb.TCSState) error {
                if err != nil {
                    sp.Op.State = pb.TCSState_Op_STATE_FAILED
                    sp.Op.ErrorMessage = err.Error()
                } else {
                    sp.Op.State = pb.TCSState_Op_STATE_DONE
                }
                return nil
            })
            if persistErr != nil {
                // TODO: Is there a better way here?
                panic(persistErr.Error())
            }
        }()
        return nil
    })
    return &pb.StartAsyncTranscodeReply{}, err
}

func (tcs *tcServer) CheckAsyncTranscode(ctx context.Context, req *pb.CheckAsyncTranscodeRequest) (*pb.CheckAsyncTranscodeReply, error) {
    fmt.Printf("CheckAsyncTranscode: %v\n", req)
    reply := &pb.CheckAsyncTranscodeReply{}
    err := tcs.s.Do(func(sp *pb.TCSState) error {
        if sp.Op == nil || sp.Op.State == pb.TCSState_Op_STATE_UNKNOWN {
            return errors.New("No active transcode")
        }
        if sp.Op.Name != req.Name {
            return fmt.Errorf("Active transcode is named '%s', but '%s' was requested", sp.Op.Name, req.Name)
        }
        reply.State = func() pb.CheckAsyncTranscodeReply_State {
            switch sp.Op.State {
            case pb.TCSState_Op_STATE_UNKNOWN:
                return pb.CheckAsyncTranscodeReply_STATE_UNKNOWN
            case pb.TCSState_Op_STATE_IN_PROGRESS:
                return pb.CheckAsyncTranscodeReply_STATE_IN_PROGRESS
            case pb.TCSState_Op_STATE_DONE:
                return pb.CheckAsyncTranscodeReply_STATE_DONE
            case pb.TCSState_Op_STATE_FAILED:
                return pb.CheckAsyncTranscodeReply_STATE_FAILED
            default:
                panic(fmt.Sprintf("Unexpected op state %v", sp.Op.State))
            }
        }()
        reply.ErrorMessage = sp.Op.ErrorMessage
        return nil
    })
    return reply, err
}

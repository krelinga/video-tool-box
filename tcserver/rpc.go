package main

import (
    "context"
    "fmt"
    "errors"
    "io"
    "os"
    "os/exec"

    "github.com/krelinga/video-tool-box/pb"
)

type tcServer struct {
    pb.UnimplementedTCServerServer

    defaultProfile  string
    s               *state
}

func newTcServer(statePath, profile string) *tcServer {
    return &tcServer{
        s: newState(statePath),
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

type handbrakeFlags []string

var gHandbrakeProfile = map[string]handbrakeFlags{
    "mkv_h265_1080p30": {
        "-Z", "Matroska/H.265 MKV 1080p30",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2160p60_surround": {
        "-Z", "General/Super HQ 2160p60 4K HEVC Surround",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2160p60_no_audio": {
        "-Z", "General/Super HQ 2160p60 4K HEVC Surround",
        "--audio", "none",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_1080p30_no_sound": {
        "-Z", "Matroska/H.265 MKV 1080p30",
        "--audio", "none",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2160p60_fast": {
        "-Z", "General/Fast 2160p60 4K HEVC",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2160p60_very_fast": {
        "-Z", "General/Very Fast 2160p60 4K HEVC",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
}

func transcodeImpl(inPath, outPath, profile string, s *state) error {
    profileFlags, ok := gHandbrakeProfile[profile]
    if !ok {
        return fmt.Errorf("unknown profile %s", profile)
    }
    standardFlags := []string{
        "-i", inPath,
        "-o", outPath,
        "--json",
    }

    if _, err := os.Stat(outPath); !errors.Is(err, os.ErrNotExist) {
        return errors.New("Output file already exists")
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

    // A pipe to allow stdout to be consumed via a Reader.
    hbPipeReader, hbPipeWriter := io.Pipe()

    // Tee the output of Handbrake so that it goes to both stdOutFile and
    // progressReader
    progressReader := io.TeeReader(hbPipeReader, stdOutFile)

    // parse entries out of progressReader and into a channel.
    progressCh := parseHbOutput(progressReader)

    // Consume from progressCh while Handbrake is running, and update s.
    // Notify progressDone when all updates have been consumed.
    progressDone := make(chan struct{})
    go func() {
        for u := range(progressCh) {
            s.Do(func(_ *pb.TCSState, prog **hbProgress) error {
                *prog = u
                return nil
            })
        }
        progressDone <- struct{}{}
    }()

    cmd := exec.Command("HandBrakeCLI")
    cmd.Args = append(cmd.Args, standardFlags...)
    cmd.Args = append(cmd.Args, profileFlags...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = hbPipeWriter
    cmd.Stderr = stdErrFile
    err = cmd.Run()

    // Wait for all progress to be processed before we exit.  This also makes
    // sure that all output from Handbrake was written to stdOutFile via the
    // above tee.  Note that the tee does not close stdOutFile when hbPipeWriter
    // is closed, so we rely on a `defer stdOutFile.Close()` statement above.
    hbPipeWriter.Close()
    <- progressDone

    return err
}

// Starts Handbrake and blocks until it finishes.
func (tcs *tcServer) transcode(inPath, outPath, profile string) {
    err := func() error {
        if err := copyRelatedFiles(inPath, outPath); err != nil {
            return err
        }
        if err := transcodeImpl(inPath, outPath, profile, tcs.s); err != nil {
            return err
        }
        return nil
    }()
    persistErr := tcs.s.Do(func(sp *pb.TCSState, prog **hbProgress) error {
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
    err := tcs.s.Do(func(sp *pb.TCSState, _ **hbProgress) error {
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
    err := tcs.s.Do(func(sp *pb.TCSState, prog **hbProgress) error {
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
        if *prog != nil {
            reply.Progress = (*prog).String()
        }
        return nil
    })
    return reply, err
}

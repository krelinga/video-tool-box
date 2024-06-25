package main

import (
    "context"
    "errors"
    "fmt"
    "strings"

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
    err := tcs.tc.StartShow(req.Name, req.InDirPath, req.OutParentDirPath, profile)
    return &pb.StartAsyncShowTranscodeReply{}, err
}

func cleanEpisode(episodePath, showDir string) string {
    suffix, found := strings.CutPrefix(episodePath, showDir)
    if !found {
        panic(episodePath)
    }
    clean, _ := strings.CutPrefix(suffix, "/")
    return clean
}

func (tcs *tcServer) CheckAsyncShowTranscode(ctx context.Context, req *pb.CheckAsyncShowTranscodeRequest) (*pb.CheckAsyncShowTranscodeReply, error) {
    fmt.Printf("CheckAsyncShowTranscode: %v\n", req)
    reply := &pb.CheckAsyncShowTranscodeReply{}
    readState := func(s *transcoder.ShowState) {
        for _, fileState := range s.FileStates {
            fileStateProto := &pb.CheckAsyncShowTranscodeReply_File{}
            reply.File = append(reply.File, fileStateProto)
            fileStateProto.Episode = cleanEpisode(fileState.InPath(), s.InDirPath())
            switch fileState.St {
            case transcoder.StateNotStarted:
                fileStateProto.State = pb.TranscodeState_NOT_STARTED
            case transcoder.StateInProgress:
                fileStateProto.State = pb.TranscodeState_IN_PROGRESS
                if fileState.Latest != nil {
                    fileStateProto.Progress = fileState.Latest.String()
                }
            case transcoder.StateComplete:
                fileStateProto.State = pb.TranscodeState_DONE
            case transcoder.StateError:
                fileStateProto.State = pb.TranscodeState_FAILED
                fileStateProto.ErrorMessage = fileState.Err.Error()
            default:
                panic(fileState.St)
            }
        }
        switch s.St {
        case transcoder.StateNotStarted:
            reply.State = pb.TranscodeState_NOT_STARTED
        case transcoder.StateInProgress:
            reply.State = pb.TranscodeState_IN_PROGRESS
        case transcoder.StateComplete:
            reply.State = pb.TranscodeState_DONE
        case transcoder.StateError:
            reply.State = pb.TranscodeState_FAILED
            reply.ErrorMessage = s.Err.Error()
        }
    }
    return reply, tcs.tc.CheckShow(req.Name, readState)
}

func (tcs *tcServer) StartAsyncSpreadTranscode(ctx context.Context, req *pb.StartAsyncSpreadTranscodeRequest) (*pb.StartAsyncSpreadTranscodeReply, error) {
    fmt.Printf("StartAsyncSpreadTranscode: %v\n", req)
    var profiles []string
    if req.ProfileList != nil {
        profiles = req.ProfileList.Profile
    } else {
        return nil, errors.New("No profile info specified")
    }
    err := tcs.tc.StartSpread(req.Name, req.InPath, req.OutParentDirPath, profiles)
    return &pb.StartAsyncSpreadTranscodeReply{}, err
}

func (tcs *tcServer) CheckAsyncSpreadTranscode(ctx context.Context, req *pb.CheckAsyncSpreadTranscodeRequest) (*pb.CheckAsyncSpreadTranscodeReply, error) {
    fmt.Printf("CheckAsyncSpreadTranscode: %v\n", req)
    reply := &pb.CheckAsyncSpreadTranscodeReply{}
    readState := func(s *transcoder.SpreadState) {
        for _, profileState := range s.FileStates {
            profileStateProto := &pb.CheckAsyncSpreadTranscodeReply_Profile{}
            reply.Profile = append(reply.Profile, profileStateProto)
            profileStateProto.Profile = profileState.Profile()
            switch profileState.St {
            case transcoder.StateNotStarted:
                profileStateProto.State = pb.TranscodeState_NOT_STARTED
            case transcoder.StateInProgress:
                profileStateProto.State = pb.TranscodeState_IN_PROGRESS
                if profileState.Latest != nil {
                    profileStateProto.Progress = profileState.Latest.String()
                }
            case transcoder.StateComplete:
                profileStateProto.State = pb.TranscodeState_DONE
            case transcoder.StateError:
                profileStateProto.State = pb.TranscodeState_FAILED
                profileStateProto.ErrorMessage = profileState.Err.Error()
            default:
                panic(profileState.St)
            }
        }
        switch s.St {
        case transcoder.StateNotStarted:
            reply.State = pb.TranscodeState_NOT_STARTED
        case transcoder.StateInProgress:
            reply.State = pb.TranscodeState_IN_PROGRESS
        case transcoder.StateComplete:
            reply.State = pb.TranscodeState_DONE
        case transcoder.StateError:
            reply.State = pb.TranscodeState_FAILED
            reply.ErrorMessage = s.Err.Error()
        }
    }
    return reply, tcs.tc.CheckSpread(req.Name, readState)
}

func (tcs *tcServer) ListAsyncTranscodes(ctx context.Context, req *pb.ListAsyncTranscodesRequest) (*pb.ListAsyncTranscodesReply, error) {
    out := &pb.ListAsyncTranscodesReply {}
    for _, op := range tcs.tc.List() {
        opProto := &pb.ListAsyncTranscodesReply_Op{
            Name: op.Name,
            Type: func() pb.ListAsyncTranscodesReply_Op_Type {
                switch op.Typ {
                case transcoder.TypeSingleFile:
                    return pb.ListAsyncTranscodesReply_Op_SINGLE_FILE
                case transcoder.TypeShow:
                    return pb.ListAsyncTranscodesReply_Op_SHOW
                case transcoder.TypeSpread:
                    return pb.ListAsyncTranscodesReply_Op_SPREAD
                default:
                    panic(op.Typ)
                }
            }(),
        }
        out.Op = append(out.Op, opProto)
    }
    return out, nil
}

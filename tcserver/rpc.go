package main

import (
    "context"
    "errors"
    "fmt"
    "strings"

    "github.com/krelinga/video-tool-box/tcserver/transcoder"

    pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/tcserver/v1"
    pbgrpc "buf.build/gen/go/krelinga/proto/grpc/go/krelinga/video/tcserver/v1/tcserverv1grpc"
)

type tcServer struct {
    pbgrpc.UnimplementedTCServiceServer

    defaultProfile  string
    tc* transcoder.Transcoder
}

func newTcServer(defaultProfile string, tc *transcoder.Transcoder) *tcServer {
    return &tcServer{
        defaultProfile: defaultProfile,
        tc: tc,
    }
}

func (tcs *tcServer) StartAsyncTranscode(ctx context.Context, req *pb.StartAsyncTranscodeRequest) (*pb.StartAsyncTranscodeResponse, error) {
    fmt.Printf("StartAsyncTranscode: %v\n", req)
    profile := func() string {
        if len(req.Profile) > 0 {
            return req.Profile
        }
        return tcs.defaultProfile
    }()
    fmt.Printf("using profile %s\n", profile)
    err := tcs.tc.StartFile(req.Name, req.InPath, req.OutPath, profile)
    return &pb.StartAsyncTranscodeResponse{}, err
}

func (tcs *tcServer) CheckAsyncTranscode(ctx context.Context, req *pb.CheckAsyncTranscodeRequest) (*pb.CheckAsyncTranscodeResponse, error) {
    fmt.Printf("CheckAsyncTranscode: %v\n", req)
    reply := &pb.CheckAsyncTranscodeResponse{}
    readState := func(s *transcoder.SingleFileState) {
        switch s.St {
        case transcoder.StateNotStarted:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_NOT_STARTED
        case transcoder.StateInProgress:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_IN_PROGRESS
            if s.Latest != nil {
                reply.Progress = s.Latest.String()
            }
        case transcoder.StateComplete:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_DONE
        case transcoder.StateError:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_FAILED
            reply.ErrorMessage = s.Err.Error()
        default:
            panic(s.St)
        }
    }
    return reply, tcs.tc.CheckFile(req.Name, readState)
}

func (tcs *tcServer) StartAsyncShowTranscode(ctx context.Context, req *pb.StartAsyncShowTranscodeRequest) (*pb.StartAsyncShowTranscodeResponse, error) {
    fmt.Printf("StartAsyncShowTranscode: %v\n", req)
    profile := func() string {
        if len(req.Profile) > 0 {
            return req.Profile
        }
        return tcs.defaultProfile
    }()
    fmt.Printf("using profile %s\n", profile)
    err := tcs.tc.StartShow(req.Name, req.InDirPath, req.OutParentDirPath, profile)
    return &pb.StartAsyncShowTranscodeResponse{}, err
}

func cleanEpisode(episodePath, showDir string) string {
    suffix, found := strings.CutPrefix(episodePath, showDir)
    if !found {
        panic(episodePath)
    }
    clean, _ := strings.CutPrefix(suffix, "/")
    return clean
}

func (tcs *tcServer) CheckAsyncShowTranscode(ctx context.Context, req *pb.CheckAsyncShowTranscodeRequest) (*pb.CheckAsyncShowTranscodeResponse, error) {
    fmt.Printf("CheckAsyncShowTranscode: %v\n", req)
    reply := &pb.CheckAsyncShowTranscodeResponse{}
    readState := func(s *transcoder.ShowState) {
        for _, fileState := range s.FileStates {
            fileStateProto := &pb.CheckAsyncShowTranscodeResponse_File{}
            reply.File = append(reply.File, fileStateProto)
            fileStateProto.Episode = cleanEpisode(fileState.InPath(), s.InDirPath())
            switch fileState.St {
            case transcoder.StateNotStarted:
                fileStateProto.State = pb.TranscodeState_TRANSCODE_STATE_NOT_STARTED
            case transcoder.StateInProgress:
                fileStateProto.State = pb.TranscodeState_TRANSCODE_STATE_IN_PROGRESS
                if fileState.Latest != nil {
                    fileStateProto.Progress = fileState.Latest.String()
                }
            case transcoder.StateComplete:
                fileStateProto.State = pb.TranscodeState_TRANSCODE_STATE_DONE
            case transcoder.StateError:
                fileStateProto.State = pb.TranscodeState_TRANSCODE_STATE_FAILED
                fileStateProto.ErrorMessage = fileState.Err.Error()
            default:
                panic(fileState.St)
            }
        }
        switch s.St {
        case transcoder.StateNotStarted:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_NOT_STARTED
        case transcoder.StateInProgress:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_IN_PROGRESS
        case transcoder.StateComplete:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_DONE
        case transcoder.StateError:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_FAILED
            reply.ErrorMessage = s.Err.Error()
        }
    }
    return reply, tcs.tc.CheckShow(req.Name, readState)
}

func (tcs *tcServer) StartAsyncSpreadTranscode(ctx context.Context, req *pb.StartAsyncSpreadTranscodeRequest) (*pb.StartAsyncSpreadTranscodeResponse, error) {
    fmt.Printf("StartAsyncSpreadTranscode: %v\n", req)
    var profiles []string
    if req.ProfileList != nil {
        profiles = req.ProfileList.Profile
    } else {
        return nil, errors.New("No profile info specified")
    }
    err := tcs.tc.StartSpread(req.Name, req.InPath, req.OutParentDirPath, profiles)
    return &pb.StartAsyncSpreadTranscodeResponse{}, err
}

func (tcs *tcServer) CheckAsyncSpreadTranscode(ctx context.Context, req *pb.CheckAsyncSpreadTranscodeRequest) (*pb.CheckAsyncSpreadTranscodeResponse, error) {
    fmt.Printf("CheckAsyncSpreadTranscode: %v\n", req)
    reply := &pb.CheckAsyncSpreadTranscodeResponse{}
    readState := func(s *transcoder.SpreadState) {
        for _, profileState := range s.FileStates {
            profileStateProto := &pb.CheckAsyncSpreadTranscodeResponse_Profile{}
            reply.Profile = append(reply.Profile, profileStateProto)
            profileStateProto.Profile = profileState.Profile()
            switch profileState.St {
            case transcoder.StateNotStarted:
                profileStateProto.State = pb.TranscodeState_TRANSCODE_STATE_NOT_STARTED
            case transcoder.StateInProgress:
                profileStateProto.State = pb.TranscodeState_TRANSCODE_STATE_IN_PROGRESS
                if profileState.Latest != nil {
                    profileStateProto.Progress = profileState.Latest.String()
                }
            case transcoder.StateComplete:
                profileStateProto.State = pb.TranscodeState_TRANSCODE_STATE_DONE
            case transcoder.StateError:
                profileStateProto.State = pb.TranscodeState_TRANSCODE_STATE_FAILED
                profileStateProto.ErrorMessage = profileState.Err.Error()
            default:
                panic(profileState.St)
            }
        }
        switch s.St {
        case transcoder.StateNotStarted:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_NOT_STARTED
        case transcoder.StateInProgress:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_IN_PROGRESS
        case transcoder.StateComplete:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_DONE
        case transcoder.StateError:
            reply.State = pb.TranscodeState_TRANSCODE_STATE_FAILED
            reply.ErrorMessage = s.Err.Error()
        }
    }
    return reply, tcs.tc.CheckSpread(req.Name, readState)
}

func (tcs *tcServer) ListAsyncTranscodes(ctx context.Context, req *pb.ListAsyncTranscodesRequest) (*pb.ListAsyncTranscodesResponse, error) {
    out := &pb.ListAsyncTranscodesResponse{}
    for _, op := range tcs.tc.List() {
        opProto := &pb.ListAsyncTranscodesResponse_Op{
            Name: op.Name,
            Type: func() pb.ListAsyncTranscodesResponse_Op_Type {
                switch op.Typ {
                case transcoder.TypeSingleFile:
                    return pb.ListAsyncTranscodesResponse_Op_TYPE_SINGLE_FILE
                case transcoder.TypeShow:
                    return pb.ListAsyncTranscodesResponse_Op_TYPE_SHOW
                case transcoder.TypeSpread:
                    return pb.ListAsyncTranscodesResponse_Op_TYPE_SPREAD
                default:
                    panic(op.Typ)
                }
            }(),
            State: func() pb.TranscodeState {
                switch op.St {
                case transcoder.StateNotStarted:
                    return pb.TranscodeState_TRANSCODE_STATE_NOT_STARTED
                case transcoder.StateInProgress:
                    return pb.TranscodeState_TRANSCODE_STATE_IN_PROGRESS
                case transcoder.StateComplete:
                    return pb.TranscodeState_TRANSCODE_STATE_DONE
                case transcoder.StateError:
                    return pb.TranscodeState_TRANSCODE_STATE_FAILED
                default:
                    panic(op.St)
                }
            }(),
        }
        out.Op = append(out.Op, opProto)
    }
    return out, nil
}

package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/krelinga/video-tool-box/tcserver/transcoder"

	pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/tcserver/v1"
)

type tcServer struct {
	defaultProfile string
	tc             *transcoder.Transcoder
}

func newTcServer(defaultProfile string, tc *transcoder.Transcoder) *tcServer {
	return &tcServer{
		defaultProfile: defaultProfile,
		tc:             tc,
	}
}

func (tcs *tcServer) StartAsyncTranscode(ctx context.Context, req *connect.Request[pb.StartAsyncTranscodeRequest]) (*connect.Response[pb.StartAsyncTranscodeResponse], error) {
	fmt.Printf("StartAsyncTranscode: %v\n", req)
	profile := func() string {
		if len(req.Msg.Profile) > 0 {
			return req.Msg.Profile
		}
		return tcs.defaultProfile
	}()
	fmt.Printf("using profile %s\n", profile)
	err := tcs.tc.StartFile(req.Msg.Name, req.Msg.InPath, req.Msg.OutPath, profile)
	return connect.NewResponse(&pb.StartAsyncTranscodeResponse{}), err
}

func (tcs *tcServer) CheckAsyncTranscode(ctx context.Context, req *connect.Request[pb.CheckAsyncTranscodeRequest]) (*connect.Response[pb.CheckAsyncTranscodeResponse], error) {
	fmt.Printf("CheckAsyncTranscode: %v\n", req)
	reply := connect.NewResponse(&pb.CheckAsyncTranscodeResponse{})
	readState := func(s *transcoder.SingleFileState) {
		reply.Msg.Profile = s.Profile()
		switch s.St {
		case transcoder.StateNotStarted:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_NOT_STARTED
		case transcoder.StateInProgress:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_IN_PROGRESS
			if s.Latest != nil {
				reply.Msg.Progress = s.Latest.String()
			}
		case transcoder.StateComplete:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_DONE
		case transcoder.StateError:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_FAILED
			reply.Msg.ErrorMessage = s.Err.Error()
		default:
			panic(s.St)
		}
	}
	return reply, tcs.tc.CheckFile(req.Msg.Name, readState)
}

func (tcs *tcServer) StartAsyncShowTranscode(ctx context.Context, req *connect.Request[pb.StartAsyncShowTranscodeRequest]) (*connect.Response[pb.StartAsyncShowTranscodeResponse], error) {
	fmt.Printf("StartAsyncShowTranscode: %v\n", req)
	profile := func() string {
		if len(req.Msg.Profile) > 0 {
			return req.Msg.Profile
		}
		return tcs.defaultProfile
	}()
	fmt.Printf("using profile %s\n", profile)
	err := tcs.tc.StartShow(req.Msg.Name, req.Msg.InDirPath, req.Msg.OutParentDirPath, profile)
	return connect.NewResponse(&pb.StartAsyncShowTranscodeResponse{}), err
}

func cleanEpisode(episodePath, showDir string) string {
	suffix, found := strings.CutPrefix(episodePath, showDir)
	if !found {
		panic(episodePath)
	}
	clean, _ := strings.CutPrefix(suffix, "/")
	return clean
}

func (tcs *tcServer) CheckAsyncShowTranscode(ctx context.Context, req *connect.Request[pb.CheckAsyncShowTranscodeRequest]) (*connect.Response[pb.CheckAsyncShowTranscodeResponse], error) {
	fmt.Printf("CheckAsyncShowTranscode: %v\n", req)
	reply := connect.NewResponse(&pb.CheckAsyncShowTranscodeResponse{})
	readState := func(s *transcoder.ShowState) {
		reply.Msg.Profile = s.Profile()
		for _, fileState := range s.FileStates {
			fileStateProto := &pb.CheckAsyncShowTranscodeResponse_File{}
			reply.Msg.File = append(reply.Msg.File, fileStateProto)
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
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_NOT_STARTED
		case transcoder.StateInProgress:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_IN_PROGRESS
		case transcoder.StateComplete:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_DONE
		case transcoder.StateError:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_FAILED
			reply.Msg.ErrorMessage = s.Err.Error()
		}
	}
	return reply, tcs.tc.CheckShow(req.Msg.Name, readState)
}

func (tcs *tcServer) StartAsyncSpreadTranscode(ctx context.Context, req *connect.Request[pb.StartAsyncSpreadTranscodeRequest]) (*connect.Response[pb.StartAsyncSpreadTranscodeResponse], error) {
	fmt.Printf("StartAsyncSpreadTranscode: %v\n", req)
	var profiles []string
	if req.Msg.ProfileList != nil {
		profiles = req.Msg.ProfileList.Profile
	} else {
		return nil, errors.New("No profile info specified")
	}
	err := tcs.tc.StartSpread(req.Msg.Name, req.Msg.InPath, req.Msg.OutParentDirPath, profiles)
	return connect.NewResponse(&pb.StartAsyncSpreadTranscodeResponse{}), err
}

func (tcs *tcServer) CheckAsyncSpreadTranscode(ctx context.Context, req *connect.Request[pb.CheckAsyncSpreadTranscodeRequest]) (*connect.Response[pb.CheckAsyncSpreadTranscodeResponse], error) {
	fmt.Printf("CheckAsyncSpreadTranscode: %v\n", req)
	reply := connect.NewResponse(&pb.CheckAsyncSpreadTranscodeResponse{})
	readState := func(s *transcoder.SpreadState) {
		for _, profileState := range s.FileStates {
			profileStateProto := &pb.CheckAsyncSpreadTranscodeResponse_Profile{}
			reply.Msg.Profile = append(reply.Msg.Profile, profileStateProto)
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
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_NOT_STARTED
		case transcoder.StateInProgress:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_IN_PROGRESS
		case transcoder.StateComplete:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_DONE
		case transcoder.StateError:
			reply.Msg.State = pb.TranscodeState_TRANSCODE_STATE_FAILED
			reply.Msg.ErrorMessage = s.Err.Error()
		}
	}
	return reply, tcs.tc.CheckSpread(req.Msg.Name, readState)
}

func (tcs *tcServer) ListAsyncTranscodes(ctx context.Context, req *connect.Request[pb.ListAsyncTranscodesRequest]) (*connect.Response[pb.ListAsyncTranscodesResponse], error) {
	out := connect.NewResponse(&pb.ListAsyncTranscodesResponse{})
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
		out.Msg.Op = append(out.Msg.Op, opProto)
	}
	return out, nil
}

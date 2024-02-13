package main

import (
    "os"
    "sync"

    "github.com/krelinga/video-tool-box/pb"
    "google.golang.org/protobuf/proto"
)

type state struct {
    path    string
    mu      sync.Mutex
}

func newState(path string) *state {
    return &state{path: path}
}

func (s *state) Do(fn func(*pb.TCSState) error) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    stateProto := &pb.TCSState{}
    data, err := os.ReadFile(s.path)
    if err != nil {
        if !os.IsNotExist(err) {
            return err
        }
        // Special case: the file doesn't exist.  It's OK to use a default
        // proto instance in this case.
    } else {
        if err := proto.Unmarshal(data, stateProto); err != nil {
            return err
        }
    }

    oldStateProto := proto.Clone(stateProto)

    if err := fn(stateProto); err != nil {
        return err
    }

    if proto.Equal(oldStateProto, stateProto) {
        // fn() didn't change stateProto, so no need to re-serialize it.
        return nil
    }

    data, err = proto.Marshal(stateProto)
    if err != nil {
        return err
    }
    return os.WriteFile(s.path, data, 0644)
}

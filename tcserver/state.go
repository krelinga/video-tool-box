package main

import (
    "sync"

    "github.com/krelinga/video-tool-box/pb"
    "github.com/krelinga/video-tool-box/tcserver/hb"
)

type state struct {
    stateProto  pb.TCSState
    progress    *hb.Progress
    mu          sync.Mutex
}

func (s *state) Do(fn func(*pb.TCSState, **hb.Progress) error) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    return fn(&s.stateProto, &s.progress)
}

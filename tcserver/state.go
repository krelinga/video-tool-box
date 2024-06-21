package main

import (
    "sync"

    "github.com/krelinga/video-tool-box/pb"
)

type state struct {
    stateProto  pb.TCSState
    progress    *hbProgress
    mu          sync.Mutex
}

func (s *state) Do(fn func(*pb.TCSState, **hbProgress) error) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    return fn(&s.stateProto, &s.progress)
}

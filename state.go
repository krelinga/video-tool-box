package main

import (
    "encoding/json"
    "os"
)

type projectType int

const (
    ptUndef projectType = iota
    ptMovie
    ptShow
)

type toolState struct {
    Pt      projectType
    Name    string
}

func loadToolState(p string) (*toolState, error) {
    bytes, err := os.ReadFile(p)
    if err != nil {
        if os.IsNotExist(err) {
            return &toolState{}, nil
        }
        return nil, err
    }
    ts := &toolState{}
    if err := json.Unmarshal(bytes, &ts); err != nil {
        return nil, err
    }
    return ts, nil
}

func (ts *toolState) store(p string) error {
    bytes, err := json.Marshal(ts)
    if err != nil { return err }
    return os.WriteFile(p, bytes, 0644)
}

var gToolState *toolState = nil

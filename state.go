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

func (pt projectType) String() string {
    switch pt {
    case ptUndef: return "ptUndef"
    case ptMovie: return "ptMovie"
    case ptShow: return "ptShow"
    }
    panic("unexpected projectType value")
}

type toolState struct {
    Pt      projectType
    Name    string
}

func readToolState(path string) (ts toolState, err error) {
    bytes, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // Special case: state file doesn't exist.
            err = nil
            return
        }
        return 
    }
    err = json.Unmarshal(bytes, &ts)
    return
}

func writeToolState(ts toolState, path string) error {
    bytes, err := json.Marshal(ts)
    if err != nil {
        return err
    }
    return os.WriteFile(path, bytes, 0644)
}

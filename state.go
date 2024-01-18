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

// Global instance of toolState shared by the whole binary.
var gToolState toolState

// read state into gToolState
func readToolState() error {
    paths, err := newToolPaths()
    if err != nil { return err }
    bytes, err := os.ReadFile(paths.StatePath())
    if err != nil {
        if os.IsNotExist(err) {
            // Special case: state file doesn't exist.
            gToolState = toolState{}
            return nil
        }
        return err
    }
    temp := toolState{}
    if err := json.Unmarshal(bytes, &temp); err != nil {
        return err
    }
    gToolState = temp
    return nil
}

// writes values from gToolState
func writeToolState() error {
    bytes, err := json.Marshal(gToolState)
    if err != nil {
        return err
    }
    paths, err := newToolPaths()
    if err != nil { return err }
    return os.WriteFile(paths.StatePath(), bytes, 0644)
}

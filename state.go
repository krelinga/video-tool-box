package main

import (
    "encoding/json"
    "fmt"
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

func (ts toolState) String() string {
    return fmt.Sprintf("toolState{Pt: %s, Name: %s}", ts.Pt, ts.Name)
}

// Global instance of toolState shared by the whole binary.
var gToolState toolState

// read state into gToolState
func readToolState() error {
    paths, err := newProdToolPaths()
    if err != nil { return err }
    ts, err := newToolState(paths.StatePath())
    if err != nil {
        return err
    }
    gToolState = ts
    return nil
}

func newToolState(path string) (ts toolState, err error) {
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

// writes values from gToolState
func writeToolState() error {
    bytes, err := json.Marshal(gToolState)
    if err != nil {
        return err
    }
    paths, err := newProdToolPaths()
    if err != nil { return err }
    return os.WriteFile(paths.StatePath(), bytes, 0644)
}

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

// Global instance of toolState shared by the whole binary.
var gToolState toolState

// read state into gToolState
func readToolState() error {
    bytes, err := os.ReadFile(statePath)
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
    return os.WriteFile(statePath, bytes, 0644)
}

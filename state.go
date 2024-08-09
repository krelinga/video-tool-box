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

type projectStage int

const (
    psUndef = iota
    psWorking
    psReadyForPush
    psPushed
)

func (ps projectStage) String() string {
    switch ps {
    case psUndef: return "psUndef"
    case psWorking: return "psWorking"
    case psReadyForPush: return "psReadyForPush"
    case psPushed: return "psPushed"
    }
    panic("unexpected projectStage value")
}

type projectState struct {
    Name string
    Stage projectStage
    Pt      projectType
    // Set if TMM has forced an override to the otherwise-computed project name.
    TmmDirOverride  string
}

type toolState struct {
    Projects []*projectState
}

func (ts *toolState) FindByName(name string) (*projectState, bool) {
    for _, ps := range ts.Projects {
        if ps.Name == name {
            return ps, true
        }
    }
    return nil, false
}

func (ts *toolState) FindByStage(s projectStage) []*projectState {
    found := []*projectState{}
    for _, p := range ts.Projects {
        if p.Stage == s {
            found = append(found, p)
        }
    }
    return found
}

func readToolState(path string) (ts *toolState, err error) {
    bytes, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // Special case: state file doesn't exist.
            ts = &toolState{}
            err = nil
            return
        }
        return 
    }
    ts = &toolState{}
    err = json.Unmarshal(bytes, ts)
    return
}

func writeToolState(ts *toolState, path string) error {
    bytes, err := json.Marshal(ts)
    if err != nil {
        return err
    }
    return os.WriteFile(path, bytes, 0644)
}

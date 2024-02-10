package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "strings"
)

type hbProgress struct {
    State       string
    Working     *hbProgressWorking
    Muxing      *hbProgressMuxing
    WorkDone    *hbProgressWorkDone
    Scanning    *hbProgressScanning
}

func (hbp *hbProgress) String() string {
    working := func() string {
        if hbp.Working == nil {
            return "nil"
        }

        return fmt.Sprintf("{%s}", hbp.Working)
    }
    muxing := func() string {
        if hbp.Muxing == nil {
            return "nil"
        }

        return fmt.Sprintf("{%s}", hbp.Muxing)
    }
    workDone := func() string {
        if hbp.WorkDone == nil {
            return "nil"
        }

        return fmt.Sprintf("{%s}", hbp.WorkDone)
    }
    scanning := func() string {
        if hbp.Scanning == nil {
            return "nil"
        }

        return fmt.Sprintf("{%s}", hbp.Scanning)
    }
    return fmt.Sprintf("State: %s Working: %s Muxing: %s WorkDone: %s Scanning: %s", hbp.State, working(), muxing(), workDone(), scanning())
}

type hbProgressWorking struct {
    ETASeconds  int
    Hours       int
    Minutes     int
    Pass        int
    PassCount   int
    PassID      int
    Paused      int
    Progress    float64
    Rate        float64
    RateAvg     float64
    Seconds     int
    SequenceID  int
}

func (w *hbProgressWorking) String() string {
    return strings.Join([]string {
        fmt.Sprintf("ETASeconds: %d", w.ETASeconds),
        fmt.Sprintf("Hours: %d", w.Hours),
        fmt.Sprintf("Minutes: %d", w.Minutes),
        fmt.Sprintf("Pass: %d", w.Pass),
        fmt.Sprintf("PassCount: %d", w.PassCount),
        fmt.Sprintf("PassID: %d", w.PassID),
        fmt.Sprintf("Paused: %d", w.Paused),
        fmt.Sprintf("Progress: %f", w.Progress),
        fmt.Sprintf("Rate: %f", w.Rate),
        fmt.Sprintf("RateAvg: %f", w.RateAvg),
        fmt.Sprintf("Seconds: %d", w.Seconds),
        fmt.Sprintf("SequenceID: %d", w.SequenceID),
    }, " ")
}

type hbProgressMuxing struct {
    Progress    float64
}

func (m *hbProgressMuxing) String() string {
    return fmt.Sprintf("Progress: %f", m.Progress)
}

type hbProgressWorkDone struct {
    Error   int
}

func (wd *hbProgressWorkDone) String() string {
    return fmt.Sprintf("Error: %d", wd.Error)
}

type hbProgressScanning struct {
    Preview         int
    PreviewCount    int
    Progress        float64
    SequenceID      int
    Title           int
    TitleCount      int
}

func (s *hbProgressScanning) String() string {
    return strings.Join([]string{
        fmt.Sprintf("Preview: %d", s.Preview),
        fmt.Sprintf("PreviewCount: %d", s.PreviewCount),
        fmt.Sprintf("Progress: %f", s.Progress),
        fmt.Sprintf("SequenceID: %d", s.SequenceID),
        fmt.Sprintf("Title: %d", s.Title),
        fmt.Sprintf("TitleCount: %d", s.TitleCount),
    }, " ")
}

// TODO: mark the output channel as consume-only.
// Returned channel is closed once hbOutput returns EOF.
func parseHbOutput(hbOutput io.Reader) chan *hbProgress {
    out := make(chan *hbProgress)
    go func() {
        scanner := bufio.NewScanner(hbOutput)
        var byteBuffer *bytes.Buffer
        for scanner.Scan() {
            line := scanner.Text()
            if byteBuffer != nil {
                // We are in a "Progress: {" stanza.
                byteBuffer.WriteString(line)
                if line == "}" {
                    // We're at the end of a "Progress: {" stanza.
                    current := &hbProgress{}
                    if err := json.Unmarshal(byteBuffer.Bytes(), current); err != nil {
                        // TODO: find a better way to signal this.
                        panic(err)
                    }
                    byteBuffer = nil
                    out <- current
                }
            } else {
                // We are not in a "Progress: {" stanza.
                if line == "Progress: {" {
                    // We are at the start of a progress stanza.
                    byteBuffer = &bytes.Buffer{}
                    byteBuffer.WriteString("{")
                } else {
                    // Discard this line of data.
                }
            }
        }
        if err := scanner.Err(); err != nil {
            // TODO: Find a better way.
            panic(err)
        }
        close(out)
    }()
    return out
}

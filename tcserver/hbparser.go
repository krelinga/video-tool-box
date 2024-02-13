package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
)

type hbProgress struct {
    State       string
    Working     *hbProgressWorking
    Muxing      *hbProgressMuxing
    WorkDone    *hbProgressWorkDone
    Scanning    *hbProgressScanning
}

func (hbp *hbProgress) String() string {
    switch {
    case hbp.Working != nil:
        return fmt.Sprintf("Working: %s", hbp.Working)
    case hbp.Muxing != nil:
        return fmt.Sprintf("Muxing: %s", hbp.Muxing)
    case hbp.WorkDone != nil:
        return fmt.Sprintf("WorkDone: %s", hbp.WorkDone)
    case hbp.Scanning != nil:
        return fmt.Sprintf("Scanning:  %s", hbp.Scanning)
    }
    return fmt.Sprintf("Unexpected state '%s'", hbp.State)
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
    etaPart := func() string {
        if w.Hours == -1 && w.Minutes == -1 && w.Seconds == -1 {
            return fmt.Sprintf("%21s", "UKNOWN")
        } else {
            return fmt.Sprintf("%4d:%02d:%02d (%7dS)", w.Hours, w.Minutes, w.Seconds, w.ETASeconds)
        }
    }()
    return fmt.Sprintf("Pass %d/%d %7.3f%%, ETA %s, %7.3f FPS (%7.3f aFPS)", w.Pass, w.PassCount, w.Progress * 100.0, etaPart, w.Rate, w.RateAvg)
}

type hbProgressMuxing struct {
    Progress    float64
}

func (m *hbProgressMuxing) String() string {
    return fmt.Sprintf("%7.3f%%", m.Progress * 100.0)
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
    return fmt.Sprintf("Preview %d/%d %7.3f%%", s.Preview, s.PreviewCount, s.Progress * 100.0)
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

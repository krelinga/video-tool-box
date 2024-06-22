package hb

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
)

type Progress struct {
    State       string
    Working     *ProgressWorking
    Muxing      *ProgressMuxing
    WorkDone    *ProgressWorkDone
    Scanning    *ProgressScanning
}

func (p *Progress) String() string {
    switch {
    case p.Working != nil:
        return fmt.Sprintf("Working: %s", p.Working)
    case p.Muxing != nil:
        return fmt.Sprintf("Muxing: %s", p.Muxing)
    case p.WorkDone != nil:
        return fmt.Sprintf("WorkDone: %s", p.WorkDone)
    case p.Scanning != nil:
        return fmt.Sprintf("Scanning:  %s", p.Scanning)
    }
    return fmt.Sprintf("Unexpected state '%s'", p.State)
}

type ProgressWorking struct {
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

func (w *ProgressWorking) String() string {
    etaPart := func() string {
        if w.Hours == -1 && w.Minutes == -1 && w.Seconds == -1 {
            return fmt.Sprintf("%21s", "UKNOWN")
        } else {
            return fmt.Sprintf("%4d:%02d:%02d (%7dS)", w.Hours, w.Minutes, w.Seconds, w.ETASeconds)
        }
    }()
    return fmt.Sprintf("Pass %d/%d %7.3f%%, ETA %s, %7.3f FPS (%7.3f aFPS)", w.Pass, w.PassCount, w.Progress * 100.0, etaPart, w.Rate, w.RateAvg)
}

type ProgressMuxing struct {
    Progress    float64
}

func (m *ProgressMuxing) String() string {
    return fmt.Sprintf("%7.3f%%", m.Progress * 100.0)
}

type ProgressWorkDone struct {
    Error   int
}

func (wd *ProgressWorkDone) String() string {
    return fmt.Sprintf("Error: %d", wd.Error)
}

type ProgressScanning struct {
    Preview         int
    PreviewCount    int
    Progress        float64
    SequenceID      int
    Title           int
    TitleCount      int
}

func (s *ProgressScanning) String() string {
    return fmt.Sprintf("Preview %d/%d %7.3f%%", s.Preview, s.PreviewCount, s.Progress * 100.0)
}

// Returned channel is closed once Output returns EOF.
func ParseOutput(Output io.Reader) <-chan *Progress {
    out := make(chan *Progress)
    go func() {
        scanner := bufio.NewScanner(Output)
        var byteBuffer *bytes.Buffer
        for scanner.Scan() {
            line := scanner.Text()
            if byteBuffer != nil {
                // We are in a "Progress: {" stanza.
                byteBuffer.WriteString(line)
                if line == "}" {
                    // We're at the end of a "Progress: {" stanza.
                    current := &Progress{}
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

package main

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "strings"

    "github.com/krelinga/video-tool-box/pb"
)

type MkvInfoServer struct {
    pb.UnimplementedMkvInfoServer
}

func (mis *MkvInfoServer) GetMkvInfo(ctx context.Context, req *pb.GetMkvInfoRequest) (*pb.GetMkvInfoReply, error) {
    fmt.Printf("Saw request: %v\n", req)
    cmd := exec.Command("mkvmerge", "-J", req.In)
    b := &bytes.Buffer{}
    cmd.Stdout = b
    cmd.Stderr = b
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("Could not run mkvmerge tool: %w", err)
    }
    mkvmergeBytes := b.Bytes()
    j, err := newMkvmergeJson(mkvmergeBytes)
    if err != nil {
        return nil, fmt.Errorf("Could not parse mkvmerge output: %w", err)
    }
    parts := []string{}
    for _, t := range j.Tracks {
        if t.Type != "audio" {
            continue
        }
        summary := fmt.Sprintf("Track ID %d: %d channels named %s language %s", t.Id, t.Properties.AudioChannels, t.Properties.TrackName, t.Properties.Language)
        parts = append(parts, summary)
    }
    reply := &pb.GetMkvInfoReply{
        Out: string(mkvmergeBytes),
        Summary: strings.Join(parts, "\n"),
    }
    return reply, nil
}



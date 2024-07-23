package main

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"
)

type config struct {
    MkvUtilServerTarget string
    TcServerTarget string
    DefaultShowTranscodeOutDir string
    DefaultMovieTranscodeOutDir string
}

func (c *config) String() string {
    q := func(s string) string {
        if len(s) == 0 {
            return "<empty>"
        }
        return fmt.Sprintf("\"%s\"", s)
    }

    b := &strings.Builder{}
    fmt.Fprintf(b, "MkvUtilServerTarget: %s\n", q(c.MkvUtilServerTarget))
    fmt.Fprintf(b, "TcServerTarget: %s\n", q(c.TcServerTarget))
    fmt.Fprintf(b, "DefaultShowTranscodeOutDir: %s\n", q(c.DefaultShowTranscodeOutDir))
    fmt.Fprintf(b, "DefaultMovieTranscodeOutDir: %s\n", q(c.DefaultMovieTranscodeOutDir))
    return b.String()
}

func readConfig(path string) (*config, error) {
    bytes, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // Special case: config file doesn't exist.
            return &config{}, nil
        }
        return nil, err
    }
    c := &config{}
    if err := json.Unmarshal(bytes, c); err != nil {
        return nil, err
    }
    return c, nil
}

func writeConfig(c *config, path string) error {
    bytes, err := json.Marshal(c)
    if err != nil {
        return err
    }
    return os.WriteFile(path, bytes, 0644)
}

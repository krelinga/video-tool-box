package main

import (
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

const mkvSuffix = ".mkv"

func stripMkvPathSuffix(p string) (string, error) {
    out := strings.TrimSuffix(p, mkvSuffix)
    if out == p {
        return "", fmt.Errorf("path %s does not end with suffix %s", p, mkvSuffix)
    }
    return out, nil
}

func matchFilesByPathPrefix(prefix string) ([]string, error) {
    dir := filepath.Dir(prefix)
    filesInDir, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }
    out := make([]string, 0)
    for _, fileInDir := range filesInDir {
        p := filepath.Join(dir, fileInDir.Name())
        if strings.HasPrefix(p, prefix) {
            out = append(out, p)
        }
    }
    return out, nil
}

func mapPaths(inPaths []string, inPrefix, outPrefix string) map[string]string {
    out := make(map[string]string)
    for _, p := range(inPaths) {
        suffix, foundPrefix := strings.CutPrefix(p, inPrefix)
        if !foundPrefix {
            panic(fmt.Sprintf("could not find prefix %s in path %s", inPrefix, p))
        }
        out[p] = strings.Join([]string{outPrefix, suffix}, "")
    }
    return out
}

func checkForExistingDest(m map[string]string) error {
    exists := func(p string) bool {
        _, err := os.Stat(p)
        return !errors.Is(err, os.ErrNotExist)
    }
    for _, out := range m {
        if exists(out) {
            return fmt.Errorf("%s already exists", out)
        }
    }
    return nil
}

func copyFilesByMapping(m map[string]string) error {
    for inPath, outPath := range m {
        cmd := exec.Command("cp", inPath, outPath)
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("Could not copy %s to %s: %w", inPath, outPath, err)
        }
    }
    return nil
}

// removes the '.mkv' suffix from inPath, finds matching files, and copies them
// to the correspondingly-transformed part of outPath.
func copyRelatedFiles(inPath, outPath string) error {
    inPrefix, err := stripMkvPathSuffix(inPath)
    if err != nil {
        return err
    }
    outPrefix, err := stripMkvPathSuffix(outPath)
    if err != nil {
        return err
    }
    inPaths, err := matchFilesByPathPrefix(inPrefix)
    if err != nil {
        return err
    }
    mapping := mapPaths(inPaths, inPrefix, outPrefix)
    if err := checkForExistingDest(mapping); err != nil {
        return err
    }
    // Remove the entry that corresponds to the input '.mkv' file ... no need to copy that.
    delete(mapping, inPath)
    return copyFilesByMapping(mapping)
}

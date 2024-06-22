package show

import (
    "path/filepath"
    "strings"
)

func OutputDir(inDir, outParentDir string) string {
    title := filepath.Base(inDir)
    return filepath.Join(outParentDir, title)
}

func MapPaths(inPaths []string, inDir, outDir string) map[string]string {
    out := make(map[string]string)
    for _, p := range inPaths {
        child, found := strings.CutPrefix(p, inDir)
        if !found {
            panic(p)
        }
        out[p] = outDir + child
    }
    return out
}

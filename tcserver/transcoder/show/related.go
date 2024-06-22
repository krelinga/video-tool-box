package show

import (
    "os"
)

func FindRelatedFiles(dir string) ([]string, error) {
    filesInDir, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }
    out := []string{}
    for _, file := range filesInDir {
        // This will exclude season & "extras" dirs.
        if !file.IsDir() {
            out = append(out, file.Name())
        }
    }
    return out, nil
}

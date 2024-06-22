package show

import (
    "os"
    "os/exec"
    "path/filepath"
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

func copyOneFile(inPath, outPath string) error {
    if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
        return err
    }
    cmd := exec.Command("cp", inPath, outPath)
    return cmd.Run()
}

func CopyFiles(pathMap map[string]string) error {
    for in, out := range pathMap {
        if err := copyOneFile(in, out); err != nil {
            return err
        }
    }
    return nil
}

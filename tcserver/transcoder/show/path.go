package show

import (
    "path/filepath"
)

func OutputDir(inDir, outParentDir string) string {
    title := filepath.Base(inDir)
    return filepath.Join(outParentDir, title)
}

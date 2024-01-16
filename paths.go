package main

import (
    "errors"
    "os"
    "path/filepath"
)

func tsPath() (string, error) {
    homeDir := os.Getenv("HOME")
    if len(homeDir) == 0 {
        return "", errors.New("could not read $HOME")
    }
    return filepath.Join(homeDir, ".vtb_state"), nil
}

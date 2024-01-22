package main

import (
    "testing"
    "os"
)

func TestSomething(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping test in short mode.")
    }
    t.Log(os.Getenv("PWD"))
}

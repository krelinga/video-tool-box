package main

import (
    "testing"
)

func TestSomething(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping test in short mode.")
    }
}

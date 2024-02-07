package main

import (
    "fmt"
    "os"
)

func getEnvVar(name string) (string, error) {
    value := os.Getenv(name)
    if len(value) == 0 {
        return "", fmt.Errorf("env var %s is not set", name)
    }
    return value, nil
}

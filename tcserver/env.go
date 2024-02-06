package main

import (
    "errors"
    "fmt"
    "os"
)

func getEnvVar(name string) (string, error) {
    value := os.Getenv(name)
    if len(value) == 0 {
        return "", errors.New(fmt.Sprintf("env var %s is not set", name))
    }
    return value, nil
}

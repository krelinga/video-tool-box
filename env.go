package main

import (
    "errors"
    "fmt"
    "os"
)

func getEnvVar(name string) (value string, err error) {
    value = os.Getenv(name)
    if len(value) == 0 {
        errStr := fmt.Sprintf("could not read %s env var", name)
        err = errors.New(errStr)
    }
    return
}

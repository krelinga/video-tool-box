package main

import (
    "fmt"
    "os"
)

func getEnvVar(name string) (value string, err error) {
    value = os.Getenv(name)
    if len(value) == 0 {
        err = fmt.Errorf("could not read %s env var", name)
    }
    return
}

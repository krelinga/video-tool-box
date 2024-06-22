package main

import (
    "fmt"
    "os"
    "strconv"
)

func getEnvVar(name string) (string, error) {
    value := os.Getenv(name)
    if len(value) == 0 {
        return "", fmt.Errorf("env var %s is not set", name)
    }
    return value, nil
}

func getEnvVarInt(name string) (int, error) {
    strVal, err := getEnvVar(name)
    if err != nil {
        return 0, err
    }
    return strconv.Atoi(strVal)
}

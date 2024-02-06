package main

import (
    "errors"
    "fmt"
    "strings"
)

func translatePath(canonPath string) (string, error) {
    oldPrefix, err := getEnvVar("VTB_TCSERVER_IN_PATH_PREFIX")
    if err != nil {
        return "", err
    }
    newPrefix, err := getEnvVar("VTB_TCSERVER_OUT_PATH_PREFIX")
    if err != nil {
        return "", err
    }

    cutPath, found := strings.CutPrefix(canonPath, oldPrefix)
    if !found {
        return "", errors.New(fmt.Sprintf("path %s does not start with prefix %s", canonPath, oldPrefix))
    }
    return newPrefix + cutPath, nil
}

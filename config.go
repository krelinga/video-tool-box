package main

import (
    "encoding/json"
    "os"
)

type config struct {
    MkvUtilServerTarget string
}

func readConfig(path string) (*config, error) {
    bytes, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // Special case: config file doesn't exist.
            return &config{}, nil
        }
        return nil, err
    }
    c := &config{}
    if err := json.Unmarshal(bytes, c); err != nil {
        return nil, err
    }
    return c, nil
}

func writeConfig(c *config, path string) error {
    bytes, err := json.Marshal(c)
    if err != nil {
        return err
    }
    return os.WriteFile(path, bytes, 0644)
}

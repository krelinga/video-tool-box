package main

import (
    "fmt"

    cli "github.com/urfave/cli/v2"
)

func cmdNew(c *cli.Context) error {
    fmt.Println("cmdNew")
    return nil
}

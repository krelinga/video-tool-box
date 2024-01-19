package main

import (
    "fmt"

    cli "github.com/urfave/cli/v2"
)

func cmdTrans() *cli.Command {
    fn := func(c *cli.Context) error {
        fmt.Println("hello transcode.")
        return nil
    }

    return &cli.Command{
        Name: "trans",
        Usage: "transcode video",
        Action: fn,
    }
}

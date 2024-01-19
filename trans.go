package main

import (
    "fmt"

    cli "github.com/urfave/cli/v2"
)

func cmdTrans() *cli.Command {
    fn := func(c *cli.Context) error {
        if len(c.String("handbrake")) > 0 {
            fmt.Println("handbrake path:", c.String("handbrake"))
        } else {
            fmt.Println("handbrake flag empty")
        }
        return nil
    }

    return &cli.Command{
        Name: "trans",
        Usage: "transcode video",
        Action: fn,
    }
}

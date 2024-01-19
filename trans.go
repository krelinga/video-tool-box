package main

import (
    "errors"

    cli "github.com/urfave/cli/v2"
)

func cmdTrans() *cli.Command {
    fn := func(c *cli.Context) error {
        handbrake := c.String("handbrake")
        if len(handbrake) == 0 {
            return errors.New("'trans' command only available when --handbrake is set")
        }
        return nil
    }

    return &cli.Command{
        Name: "trans",
        Usage: "transcode video",
        Action: fn,
    }
}

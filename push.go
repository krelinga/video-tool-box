package main

import (
    "errors"

    cli "github.com/urfave/cli/v2"
)

func cmdCfgPush() *cli.Command {
    return &cli.Command{
        Name: "push",
        Usage: "push files from Tiny Media Manager directory to NAS.",
        Action: cmdPush,
    }
}

func cmdPush(c *cli.Context) error {
    return errors.New("'push' command is not implemented.")
}

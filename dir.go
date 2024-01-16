package main

import (
    "errors"

    cli "github.com/urfave/cli/v2"
)

func cmdDir() *cli.Command{
    fn := func(c *cli.Context) error {
        if gToolState.Pt == ptUndef {
            return errors.New("no active project")
        }
        return nil
    }

    return &cli.Command{
        Name: "dir",
        Usage: "process .mkv files in current directory one at a time",
        ArgsUsage: " ",  // Makes help text a bit nicer
        Description: "Requires and existing project.",
        Action: fn,
    }
}

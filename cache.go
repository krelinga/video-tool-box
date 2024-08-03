package main

import (
    cli "github.com/urfave/cli/v2"
)

func subcmdCfgCache() *cli.Command {
    return &cli.Command{
        Name: "cache",
        Usage: "Manipulate the server-side rip cache.",
        Subcommands: []*cli.Command{
            cmdCfgSync(),
            cmdCfgClear(),
        },
    }
}

func cmdCfgSync() *cli.Command {
    return &cli.Command{
        Name: "sync",
        Usage: "Sync server-side cache to laptop.",
        Action: cmdSync,
    }
}

func cmdSync(c *cli.Context) error {
    return nil
}

func cmdCfgClear() *cli.Command {
    return &cli.Command{
        Name: "clear",
        Usage: "Clear server-side cache.",
        Action: cmdClear,
    }
}

func cmdClear(c *cli.Context) error {
    return nil
}

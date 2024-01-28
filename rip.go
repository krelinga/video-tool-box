package main

import cli "github.com/urfave/cli/v2"

func subcmdCfgRip() *cli.Command {
    return &cli.Command{
        Name: "rip",
        Usage: "process MKV files from ripping",
        Subcommands: []*cli.Command{
            cmdCfgNew(),
            cmdCfgFinish(),
            cmdCfgDir(),
            cmdCfgMeta(),
        },
    }
}

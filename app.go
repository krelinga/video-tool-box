package main

import cli "github.com/urfave/cli/v2"

func appCfg() *cli.App {
    return &cli.App{
        Name: "vtb",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "handbrake",
                Usage: "path to handbrake binary for transcoding.",
            },
        },
        Commands: []*cli.Command{
            subcmdCfgRip(),
            cmdCfgTrans(),
        },
        // Caller should set these.
        Reader: nil,
        Writer: nil,
        ErrWriter: nil,
    }
}

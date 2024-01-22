package main

import (
    "os"

    cli "github.com/urfave/cli/v2"
)

func appMain() error {
    app := &cli.App{
        Name: "vtb",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "handbrake",
                Usage: "path to handbrake binary for transcoding.",
            },
        },
        Commands: []*cli.Command{
            cmdNew(),
            cmdFinish(),
            cmdDir(),
            cmdMeta(),
            cmdTrans(),
        },
    }
    return app.Run(os.Args)
}

package main

import (
    "context"
    cli "github.com/urfave/cli/v2"
)

func appMain(args []string) error {
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
    tp, err := newProdToolPaths()
    if err != nil {
        return err
    }
    ctx := newToolPathsContext(context.Background(), tp)
    return app.RunContext(ctx, args)
}

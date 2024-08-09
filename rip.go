package main

import (
    "errors"

    cli "github.com/urfave/cli/v2"
)

func subcmdCfgRip() *cli.Command {
    return &cli.Command{
        Name: "rip",
        Usage: "process MKV files from ripping",
        Subcommands: []*cli.Command{
            cmdCfgNew(),
            cmdCfgFinish(),
            cmdCfgDir(),
            cmdCfgTmm(),
            cmdCfgPush(),
            cmdCfgStage(),
            cmdCfgRipLs(),
        },
    }
}

// Returns toolPaths, toolState, a function to call to save updated tool state, and any error.
func ripCmdInit(c *cli.Context) (*toolPaths, *toolState, func() error, error) {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return nil, nil, nil, errors.New("toolPaths not present in context")
    }
    ts, err := readToolState(tp.StatePath())
    if err != nil {
        return nil, nil, nil, err
    }
    save := func() error {
        return writeToolState(ts, tp.StatePath())
    }

    return tp, ts, save, nil
}

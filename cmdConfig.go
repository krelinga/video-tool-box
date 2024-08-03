package main

import (
    "errors"
    "fmt"

    cli "github.com/urfave/cli/v2"
)

func cmdCfgConfig() *cli.Command {
    return &cli.Command{
        Name: "config",
        Usage: "show or modify config",
        Description: "show or modify config",
        Action: cmdConfig,
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "MkvUtilServerTarget",
                Usage: "set RPC target for Mkv Util Server.",
            },
            &cli.StringFlag{
                Name: "TcServerTarget",
                Usage: "set RPC target for Transcode Server.",
            },
            &cli.StringFlag{
                Name: "DefaultShowTranscodeOutDir",
                Usage: "set default show transcode output dir.",
            },
            &cli.StringFlag{
                Name: "DefaultMovieTranscodeOutDir",
                Usage: "set default show transcode output dir.",
            },
            &cli.StringFlag{
                Name: "RipCacheDir",
                Usage: "set server-side rip cache dir.",
            },
        },
    }
}

func cmdConfig(c *cli.Context) error {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return errors.New("toolPaths not present in context")
    }
    cfg, err := readConfig(tp.ConfigPath())
    if err != nil {
        return err
    }

    needWrite := false
    if t := c.String("MkvUtilServerTarget"); len(t) > 0 {
        needWrite = true
        cfg.MkvUtilServerTarget = t
    }
    if t := c.String("TcServerTarget"); len(t) > 0 {
        needWrite = true
        cfg.TcServerTarget = t
    }
    if t := c.String("DefaultShowTranscodeOutDir"); len(t) > 0 {
        needWrite = true
        cfg.DefaultShowTranscodeOutDir = t
    }
    if t := c.String("DefaultMovieTranscodeOutDir"); len(t) > 0 {
        needWrite = true
        cfg.DefaultMovieTranscodeOutDir = t
    }
    if t := c.String("RipCacheDir"); len(t) > 0 {
        needWrite = true
        cfg.RipCacheDir = t
    }
    if needWrite {
        if err := writeConfig(cfg, tp.ConfigPath()); err != nil {
            return err
        }
    }
    _, err = fmt.Fprintf(c.App.Writer, "%s", cfg)
    return err
}

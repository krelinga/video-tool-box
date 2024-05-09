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
    if needWrite {
        if err := writeConfig(cfg, tp.ConfigPath()); err != nil {
            return err
        }
    }
    _, err = fmt.Fprintf(c.App.Writer, "%s", cfg)
    return err
}

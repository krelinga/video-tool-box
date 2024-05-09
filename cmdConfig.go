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
    _, err = fmt.Fprintf(c.App.Writer, "%s", cfg)
    return err
}

package main

import (
    "os/exec"

    cli "github.com/urfave/cli/v2"
)

func cmdCfgTmm() *cli.Command {
    return &cli.Command{
        Name: "tmm",
        Usage: "run Tiny Media Manager",
        Action: cmdTmm,
    }
}

func runTmmAndWait() error {
    cmd := exec.Command("open", "-W", "/Applications/tinyMediaManager.app/")
    return cmd.Run()
}

func cmdTmm(c *cli.Context) error {
    // TODO: record the contents of TMM directorys before & after running TMM, update project dir accordingly.

    return runTmmAndWait()
}

package main

import (
    "log"
    "os"

    cli "github.com/urfave/cli/v2"
)

func main() {
    app := &cli.App{
        Name: "vtb",
        Commands: []*cli.Command{
            cmdNew(),
            cmdFinish(),
            cmdDir(),
            cmdMeta(),
        },
    }
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}

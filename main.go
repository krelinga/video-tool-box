package main

import (
    "log"
    "os"

    cli "github.com/urfave/cli/v2"
)

func main() {
    if err := readToolState() ; err != nil {
        log.Fatal(err)
    }

    // Command line processing.
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

    // We don't defer this because we don't want to update the
    // serialized toolState if the program panics.
    if err := writeToolState(); err != nil {
        log.Fatal(err)
    }
}

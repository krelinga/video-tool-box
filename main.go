package main

import (
    "fmt"
    "log"
    "os"

    cli "github.com/urfave/cli/v2"
)

func main() {
    fmt.Println("Hello from main!")

    // Load & (eventually) store gToolState
    func() {
        var err error = nil
        gToolState, err = loadToolState(statePath)
        if err != nil {
            log.Fatal(err)
        }
    }()
    defer func() {
        if err := gToolState.store(statePath); err != nil {
            log.Fatal(err)
        }
    }()

    // Command line processing.
    app := &cli.App{
        Name: "vtb",
        Commands: []*cli.Command{
            cmdNew(),
            cmdFinish(),
            cmdDir(),
        },
    }
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}

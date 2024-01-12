package main

import (
    "fmt"
    "log"
    "os"

    cli "github.com/urfave/cli/v2"
)

func main() {
    fmt.Println("Hello from main!")
    app := &cli.App{
        Name: "vtb",
        Action: func(*cli.Context) error {
            fmt.Println("main action.")
            return nil
        },
    }
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}

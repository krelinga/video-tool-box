package main

import (
    "fmt"
    "log"
    "os"

    cli "github.com/urfave/cli/v2"
)

func listDir(dir string) {
    fmt.Println("listing dir", dir)
    entries, err := os.ReadDir(dir)
    if err != nil {
        log.Fatal(err)
    }

    for _, entry := range entries {
        dirPart := ""
        if entry.IsDir() {
            dirPart = "/"
        }
        fmt.Printf("* %s%s\n", entry.Name(), dirPart)
    }
}

func main() {
    fmt.Println("Hello from main!")
    app := &cli.App{
        Name: "vtb",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "work_dir",
                Value: "/",
                Usage: "Directory to do work in.",
            },
        },
        Action: func(c *cli.Context) error {
            fmt.Println("main action.")
            listDir(c.String("work_dir"))
            return nil
        },
    }
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}

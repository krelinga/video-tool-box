package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"

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
            work_dir := c.String("work_dir")
            listDir(work_dir)
            oldP := filepath.Join(work_dir, "moveme")
            newP := filepath.Join(work_dir, "tmm_shows", "moveme")
            os.Rename(oldP, newP)

            return nil
        },
    }
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}

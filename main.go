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
            &cli.StringFlag{
                Name: "tmm_movies",
                Value: "",
                Usage: "Tiny Media Manager movies dir.",
            },
            &cli.StringFlag{
                Name: "tmm_shows",
                Value: "",
                Usage: "Tiny Media Manager shows dir.",
            },
        },
        Action: func(c *cli.Context) error {
            fmt.Println("main action.")
            listDir(c.String("work_dir"))
            if len(c.String("tmm_movies")) > 0 {
                listDir(c.String("tmm_movies"))
            }
            if len(c.String("tmm_shows")) > 0 {
                listDir(c.String("tmm_shows"))
            }
            return nil
        },
    }
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}

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
            work_dir := c.String("work_dir")
            listDir(work_dir)
            tmm_movies := c.String("tmm_movies")
            if len(tmm_movies) > 0 {
                listDir(tmm_movies)
            }
            tmm_shows := c.String("tmm_shows")
            if len(tmm_shows) > 0 {
                listDir(tmm_shows)
            }

            return nil
        },
    }
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}

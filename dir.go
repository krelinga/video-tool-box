package main

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    cli "github.com/urfave/cli/v2"
)

func listMkvFilePaths() ([]string, error) {
    entries, err := os.ReadDir(currentDir)
    if err != nil {
        return nil, err
    }
    paths := make([]string, 0, len(entries))
    for _, entry := range entries {
        path := filepath.Join(currentDir, entry.Name())
        if filepath.Ext(path) != ".mkv" {
            continue
        }
        paths = append(paths, path)
    }
    return paths, nil
}

func cmdDir() *cli.Command{
    fn := func(c *cli.Context) error {
        if gToolState.Pt == ptUndef {
            return errors.New("no active project")
        }

        paths, err := listMkvFilePaths()
        if err != nil {
            return err
        }
        for _, path := range paths {
            fmt.Println(path)
        }

        return nil
    }

    return &cli.Command{
        Name: "dir",
        Usage: "process .mkv files in current directory one at a time",
        ArgsUsage: " ",  // Makes help text a bit nicer
        Description: "Requires and existing project.",
        Action: fn,
    }
}

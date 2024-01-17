package main

import (
    "bufio"
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

type cmdDirEarlyExit struct {}

func (_ *cmdDirEarlyExit) Error() string {
    return "interPromptEarlyExit"
}

func openInVLC(path string) error {
    fmt.Println("will open in vlc", path)
    // TODO
    return nil
}

func moveToTMMDir(path string) error {
    fmt.Println("will move to TMM dir", path)
    // TODO
    return nil
}

func moveToExtrasDir(path string) error {
    fmt.Println("will move to extras dir", path)
    // TODO
    return nil
}

func deletePath(path string) error {
    fmt.Println("will delete path", path)
    // TODO
    return nil
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

        scanner := bufio.NewScanner(os.Stdin)
        prompt := func() (string, error) {
            fmt.Println("(o)pen, (t)itle, e(x)tra, (s)kip, (d)elete, (q)uit")
            if !scanner.Scan() {
                return "", scanner.Err()
            }
            return scanner.Text(), nil
        }

        pathLoop: for _, path := range paths {
            fmt.Println(path)
            inputLoop: for {
                in, err := prompt()
                if err != nil {
                    return err
                }
                switch in {
                case "o":
                    if err := openInVLC(path); err != nil {
                        return err
                    }
                    // Repeat inputLoop
                case "t":
                    if err := moveToTMMDir(path); err != nil {
                        return err
                    }
                    break inputLoop
                case "x":
                    if err := moveToExtrasDir(path); err != nil {
                        return err
                    }
                    break inputLoop
                case "s":
                    continue pathLoop
                case "d":
                    if err := deletePath(path); err != nil {
                        return err
                    }
                    break inputLoop
                case "q":
                    break pathLoop
                }
            }
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

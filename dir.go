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
        pathLoop: for _, path := range paths {
            fmt.Println(path)
            err := interPrompt([]*interChoice{
                {
                    Text: "(o)pen",
                    Key: 'o',
                    Fn: func() error {
                        return openInVLC(path)
                    },
                },
                {
                    Text: "(t)itle",
                    Key: 't',
                    Fn: func() error {
                        return moveToTMMDir(path)
                    },
                },
                {
                    Text: "e(x)tra",
                    Key: 'x',
                    Fn: func() error {
                        return moveToExtrasDir(path)
                    },
                },
                {
                    Text: "(s)kip",
                    Key: 's',
                    Fn: func() error {
                        return nil
                    },
                },
                {
                    Text: "(d)elete",
                    Key: 'd',
                    Fn: func() error {
                        return deletePath(path)
                    },
                },
                {
                    Text: "(q)uit",
                    Key: 'q',
                    Fn: func() error {
                        return &cmdDirEarlyExit{}
                    },
                },
            })
            if err != nil {
                switch err.(type) {
                case *cmdDirEarlyExit:
                    break pathLoop
                default:
                    return err
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

package main

import (
    "bufio"
    "errors"
    "fmt"
    "math/big"
    "os"
    "os/exec"
    "path/filepath"

    cli "github.com/urfave/cli/v2"
    humanize "github.com/dustin/go-humanize"
)

func listMkvFilePaths(currentDir string) ([]string, error) {
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

func openInVLC(path string) error {
    cmd := exec.Command("open", "-a", "/Applications/VLC.app", path)
    return cmd.Run()
}

func createDestDirAndMove(toMove string, destDir string) error {
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return err
    }
    basename := filepath.Base(toMove)
    return os.Rename(toMove, filepath.Join(destDir, basename))
}

func readableFileSize(path string) (string, error) {
    info, err := os.Stat(path)
    if err != nil {
        return "", err
    }
    bigSize := big.NewInt(info.Size())
    return humanize.BigIBytes(bigSize), nil
}

func cmdCfgDir() *cli.Command {
    return &cli.Command{
        Name: "dir",
        Usage: "process .mkv files in current directory one at a time",
        ArgsUsage: " ",  // Makes help text a bit nicer
        Description: "Requires and existing project.",
        Action: cmdDir,
    }
}

func cmdDir(c *cli.Context) error {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return errors.New("missing toolPaths in context.")
    }
    ts, err := readToolState(tp.StatePath())
    if err != nil {
        return err
    }
    if ts.Pt == ptUndef {
        return errors.New("no active project")
    }

    paths, err := listMkvFilePaths(tp.CurrentDir())
    if err != nil {
        return err
    }

    scanner := bufio.NewScanner(os.Stdin)
    prompt := func() (string, error) {
        fmt.Printf("(o)pen, (t)itle, e(x)tra, (s)kip, (d)elete, (q)uit: ")
        if !scanner.Scan() {
            return "", scanner.Err()
        }
        return scanner.Text(), nil
    }
    printPath := func(path string) error {
        size, err := readableFileSize(path)
        if err != nil {
            return err
        }
        fmt.Printf("\n%s: %s\n", filepath.Base(path), size)
        return nil
    }

    pathLoop: for _, path := range paths {
        if err := printPath(path); err != nil {
            return err
        }
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
                fmt.Println("opened in VLC")
                // Repeat inputLoop
            case "t":
                destDir, err := tp.TmmProjectDir(ts)
                if err != nil { return err }
                if err := createDestDirAndMove(path, destDir); err != nil {
                    return err
                }
                fmt.Println("moved to TMM content dir")
                break inputLoop
            case "x":
                destDir, err := tp.TmmProjectExtrasDir(ts)
                if err != nil { return err }
                if err := createDestDirAndMove(path, destDir); err != nil {
                    return err
                }
                fmt.Println("moved to extras dir")
                break inputLoop
            case "s":
                fmt.Println("skipped")
                continue pathLoop
            case "d":
                if err := os.Remove(path); err != nil {
                    return err
                }
                fmt.Println("deleted")
                break inputLoop
            case "q":
                fmt.Println("quit")
                break pathLoop
            }
        }
    }

    return nil
}

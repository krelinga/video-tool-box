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
    uuid "github.com/google/uuid"
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
    exists := func(path string) error {
        _, err := os.Stat(path)
        if os.IsNotExist(err) {
            return nil
        }
        if err != nil {
            return err
        }
        return fmt.Errorf("path %s already exists", path)
    }
    basename := filepath.Base(toMove)
    destPath := filepath.Join(destDir, uuid.NewString() + "-" + basename)
    if err := exists(destPath); err != nil {
        return err
    }
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return err
    }
    return os.Rename(toMove, destPath)
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
        Usage: "process .mkv files in a directory one at a time",
        ArgsUsage: "<dir, or pwd by default>",  // Makes help text a bit nicer
        Description: "Requires an existing project.",
        Action: cmdDir,
    }
}

func cmdDir(c *cli.Context) error {
    tp, ts, _, err := ripCmdInit(c)
    if err != nil {
        return err
    }
    if ts.Pt == ptUndef {
        return errors.New("no active project")
    }

    rootDir, err := func() (string, error) {
        args := c.Args().Slice()
        switch len(args) {
        case 0:
            return tp.CurrentDir(), nil
        case 1:
            return filepath.Abs(args[0])
        default:
            return "", errors.New("only zero or one arguments supported.")
        }
    }()
    if err != nil {
        return err
    }

    paths, err := listMkvFilePaths(rootDir)
    if err != nil {
        return err
    }

    scanner := bufio.NewScanner(c.App.Reader)
    prompt := func() (string, error) {
        fmt.Fprintf(c.App.Writer, "(o)pen, (t)itle, e(x)tra, (s)kip, (d)elete, (q)uit: ")
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
        fmt.Fprintf(c.App.Writer, "\n%s: %s\n", filepath.Base(path), size)
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
                fmt.Fprintln(c.App.Writer, "opened in VLC")
                // Repeat inputLoop
            case "t":
                destDir, err := tp.TmmProjectDir(ts)
                if err != nil { return err }
                if err := createDestDirAndMove(path, destDir); err != nil {
                    return err
                }
                fmt.Fprintln(c.App.Writer, "moved to TMM content dir")
                break inputLoop
            case "x":
                destDir, err := tp.TmmProjectExtrasDir(ts)
                if err != nil { return err }
                if err := createDestDirAndMove(path, destDir); err != nil {
                    return err
                }
                fmt.Fprintln(c.App.Writer, "moved to extras dir")
                break inputLoop
            case "s":
                fmt.Fprintln(c.App.Writer, "skipped")
                continue pathLoop
            case "d":
                if err := os.Remove(path); err != nil {
                    return err
                }
                fmt.Fprintln(c.App.Writer, "deleted")
                break inputLoop
            case "q":
                fmt.Fprintln(c.App.Writer, "quit")
                break pathLoop
            }
        }
    }

    ripDirEmpty := func() bool {
        entries, err := os.ReadDir(rootDir)
        if err != nil {
            // Just swallow it ... this is only an optimization.
            return false
        }
        return len(entries) == 0
    }()
    if rootDir != tp.CurrentDir() && ripDirEmpty {
        fmt.Fprintf(c.App.Writer, "rip dir %s empty, delete it (y/N)? ", rootDir)
        var confirm string
        fmt.Fscanf(c.App.Reader, "%s", &confirm)
        if confirm != "y" {
            return nil
        }
        return os.Remove(rootDir)
    }

    return nil
}

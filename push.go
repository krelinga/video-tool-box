package main

import (
    "errors"
    "fmt"
    "io/fs"
    "os"
    "os/exec"
    "path/filepath"

    cli "github.com/urfave/cli/v2"
    humanize "github.com/dustin/go-humanize"
)

func cmdCfgPush() *cli.Command {
    return &cli.Command{
        Name: "push",
        Usage: "push files from Tiny Media Manager directory to NAS.",
        Action: cmdPush,
    }
}

func dirBytes(path string) (int64, error) {
    total := int64(0)
    walkFn := func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        info, err := d.Info()
        if err != nil {
            return err
        }
        total += info.Size()
        return nil
    }
    if err := fs.WalkDir(os.DirFS(path), ".", walkFn); err != nil {
        return 0, err
    }
    return total, nil
}

func cmdPush(c *cli.Context) error {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return errors.New("toolPaths not present in context")
    }
    ts, err := readToolState(tp.StatePath())
    if err != nil {
        return err
    }

    projectDir, err := tp.TmmProjectDir(ts)
    if err != nil {
        return err
    }

    projectDirSize, err := dirBytes(projectDir)
    if err != nil {
        return err
    }

    nasSubDir, err := func() (string, error) {
        switch ts.Pt {
        case ptUndef:
            return "", errors.New("Not in a rip project.")
        case ptMovie:
            return "Movies", nil
        case ptShow:
            return "Shows", nil
        default:
            return "", fmt.Errorf("Unexpected ProjectType value %v", ts.Pt)
        }
    }()
    if err != nil {
        return err
    }
    title := filepath.Base(projectDir)
    outSuperPath := filepath.Join(tp.NasMountDir(), nasSubDir)
    outPath := filepath.Join(outSuperPath, title)

    fmt.Fprintf(c.App.Writer, "Will copy %s from %s to %s.\nConfirm (y/N)? ", humanize.IBytes(uint64(projectDirSize)), projectDir, outPath)
    var confirm string
    fmt.Fscanf(c.App.Reader, "%s", &confirm)
    if confirm != "y" {
        return nil
    }

    // Use rsync to copy the files.
    args := []string{
        "-ah",
        "--progress",
        "-r",
        projectDir,
        outSuperPath,
    }
    cmd := exec.Command("/usr/bin/rsync", args...)
    cmd.Stdin = c.App.Reader
    cmd.Stdout = c.App.Writer
    cmd.Stderr = c.App.ErrWriter

    if err := cmd.Run(); err != nil {
        return err
    }

    // Now rename the .extras dir (if it exists)
    extrasPath := filepath.Join(outPath, ".extras")
    _, err = os.Stat(extrasPath)
    if err != nil {
        if errors.Is(err, fs.ErrNotExist) {
            fmt.Fprintln(c.App.Writer, "No extras dir.")
            return nil
        } else {
            return err
        }
    }
    newExtrasPath := filepath.Join(outPath, "extras")
    fmt.Fprintln(c.App.Writer, "Renaming extras dir.")
    return os.Rename(extrasPath, newExtrasPath)
}

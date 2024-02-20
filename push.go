package main

import (
    "errors"
    "fmt"
    "os/exec"
    "path/filepath"

    cli "github.com/urfave/cli/v2"
)

func cmdCfgPush() *cli.Command {
    return &cli.Command{
        Name: "push",
        Usage: "push files from Tiny Media Manager directory to NAS.",
        Action: cmdPush,
    }
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

    cwd := tp.CurrentDir()
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
    title := filepath.Base(cwd)
    outPath := filepath.Join(tp.NasMountDir(), nasSubDir, title)

    fmt.Fprintf(c.App.Writer, "Will copy %s to %s.\nConfirm (y/N)? ", cwd, outPath)
    var confirm string
    fmt.Fscanf(c.App.Reader, "%s", &confirm)
    if confirm != "y" {
        return nil
    }

    cmd := exec.Command("/usr/bin/rsync", "-ah", "--progress", "-r", cwd, outPath)
    cmd.Stdin = c.App.Reader
    cmd.Stdout = c.App.Writer
    cmd.Stderr = c.App.ErrWriter

    return cmd.Run()
}

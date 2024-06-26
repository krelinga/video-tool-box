package main

import (
    "errors"
    "fmt"
    "os"
    "strings"

    cli "github.com/urfave/cli/v2"
)

func cmdCfgNew() *cli.Command {
    return &cli.Command{
        Name: "new",
        Usage: "create a new project",
        ArgsUsage: "(movie|show) <name>",
        Description: "Creates a new movie or tv show project.",
        Action: cmdNew,
    }
}

func cmdNew(c *cli.Context) error {
    tp, ts, save, err := ripCmdInit(c)
    if err != nil {
        return err
    }
    if ts.Pt != ptUndef {
        return fmt.Errorf("Existsing project %s", ts.Name)
    }
    args := c.Args().Slice()
    if len(args) < 2 {
        return errors.New("Expected two arguments: project type & name")
    }
    newPt := func() projectType {
        switch args[0] {
        case "movie":
            return ptMovie
        case "show":
            return ptShow
        default:
            return ptUndef
        }
    }()
    if newPt == ptUndef {
        return errors.New("Unrecognized project type")
    }
    newName := strings.Join(args[1:], " ")
    ts.Pt = newPt
    ts.Name = newName

    projectDir, err := tp.TmmProjectDir(ts)
    if err != nil {
        return err
    }
    if err := os.MkdirAll(projectDir, 0755); err != nil {
        return fmt.Errorf("Could not create project dir %s: %w", projectDir, err)
    }

    return save()
}

func cmdCfgFinish() *cli.Command {
    return &cli.Command{
        Name: "finish",
        Usage: "finish an existing project",
        ArgsUsage: " ",  // Makes help text a bit nicer.
        Action: cmdFinish,
        Flags: []cli.Flag{
            &cli.BoolFlag{
                Name: "y",
                Usage: "always confirms interactive prompts",
                Value: false,
            },
        },
    }
}

func cmdFinish(c *cli.Context) error {
    tp, ts, save, err := ripCmdInit(c)
    if err != nil {
        return err
    }
    projectDir, err := tp.TmmProjectDir(ts)
    if err != nil {
        return err
    }
    if !c.Bool("y") {
        fmt.Fprintf(c.App.Writer, "Will delete %s.\nConfirm (y/N)? ", projectDir)
        var confirm string
        fmt.Fscanf(c.App.Reader, "%s", &confirm)
        if confirm != "y" {
            return nil
        }
    }

    if err := os.RemoveAll(projectDir); err != nil {
        return fmt.Errorf("Could not remove %s: %w", projectDir, err)
    }
    *ts = toolState{}
    return save()
}

func cmdCfgMeta() *cli.Command {
    return &cli.Command{
        Name: "meta",
        Usage: "display information about the current project",
        ArgsUsage: " ",  // Makes help text a bit nicer.
        Action: cmdMeta,
    }
}

func cmdMeta(c *cli.Context) error {
    _, ts, _, err := ripCmdInit(c)
    if err != nil {
        return err
    }
    if ts.Pt == ptUndef {
        fmt.Fprintln(c.App.Writer, "no project configured.")
        return nil
    }

    fmt.Fprintln(c.App.Writer, "Active Project")
    fmt.Fprintln(c.App.Writer, "--------------")
    fmt.Fprintln(c.App.Writer, "name:", ts.Name)
    fmt.Fprintln(c.App.Writer, "type:", ts.Pt)
    if len(ts.TmmDirOverride) > 0 {
        fmt.Fprintln(c.App.Writer, "TMM dir override:", ts.TmmDirOverride)
    }
    return nil
}

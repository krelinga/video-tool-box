package main

import (
    "errors"
    "fmt"
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
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return errors.New("toolPaths not present in context")
    }
    ts, err := readToolState(tp.StatePath())
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

    return writeToolState(ts, tp.StatePath())
}

func cmdCfgFinish() *cli.Command {
    return &cli.Command{
        Name: "finish",
        Usage: "finish an existing project",
        ArgsUsage: " ",  // Makes help text a bit nicer.
        Action: cmdFinish,
    }
}

func cmdFinish(c *cli.Context) error {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return errors.New("toolPaths not present in context")
    }
    return writeToolState(toolState{}, tp.StatePath())
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
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return errors.New("toolPaths not present in context")
    }
    ts, err := readToolState(tp.StatePath())
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
    return nil
}

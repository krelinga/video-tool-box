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
        msg := fmt.Sprintf("Existsing project %s", ts.Name)
        return errors.New(msg)
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

func cmdFinish() *cli.Command {
    fn := func(c *cli.Context) error {
        tp, ok := toolPathsFromContext(c.Context)
        if !ok {
            return errors.New("toolPaths not present in context")
        }
        return writeToolState(toolState{}, tp.StatePath())
    }

    return &cli.Command{
        Name: "finish",
        Usage: "finish an existing project",
        ArgsUsage: " ",  // Makes help text a bit nicer.
        Action: fn,
    }
}

func cmdMeta() *cli.Command {
    fn := func(c *cli.Context) error {
        tp, ok := toolPathsFromContext(c.Context)
        if !ok {
            return errors.New("toolPaths not present in context")
        }
        ts, err := readToolState(tp.StatePath())
        if err != nil {
            return err
        }
        if ts.Pt == ptUndef {
            fmt.Println("no project configured.")
            return nil
        }

        fmt.Println("Active Project")
        fmt.Println("--------------")
        fmt.Println("name:", ts.Name)
        fmt.Println("type:", ts.Pt)
        return nil
    }

    return &cli.Command{
        Name: "meta",
        Usage: "display information about the current project",
        ArgsUsage: " ",  // Makes help text a bit nicer.
        Action: fn,
    }
}

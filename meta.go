package main

import (
    "errors"
    "fmt"
    "strings"

    cli "github.com/urfave/cli/v2"
)

func cmdNew() *cli.Command {
    const argsUsage = "(movie|show) <name>"
    argsErr := errors.New(fmt.Sprintf("Usage: vtb new %s", argsUsage))

    fn := func(c *cli.Context) error {
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
            return argsErr
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
            return argsErr
        }
        newName := strings.Join(args[1:], " ")
        ts.Pt = newPt
        ts.Name = newName

        return writeToolState(ts, tp.StatePath())
    }

    return &cli.Command{
        Name: "new",
        Usage: "create a new project",
        ArgsUsage: argsUsage,
        Description: "Creates a new movie or tv show project.",
        Action: fn,
    }
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

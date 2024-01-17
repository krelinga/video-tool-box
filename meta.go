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
        if gToolState.Pt != ptUndef {
            msg := fmt.Sprintf("Existsing project %s", gToolState.Name)
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
        gToolState.Pt = newPt
        gToolState.Name = newName

        return nil
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
        gToolState = toolState{}
        return nil
    }

    return &cli.Command{
        Name: "finish",
        Usage: "finish an existing project",
        ArgsUsage: " ",  // Makes help text a bit nicer.
        Action: fn,
    }
}

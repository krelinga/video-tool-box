package main

import (
    "errors"
    "fmt"
    "path/filepath"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    cli "github.com/urfave/cli/v2"
    muspb "github.com/krelinga/mkv-util-server/pb"
)

func subcmdCfgMkv() *cli.Command {
    return &cli.Command{
        Name: "mkv",
        Usage: "Manipulate MKV files",
        Subcommands: []*cli.Command{
            cmdCfgMkvInfo(),
            cmdCfgMkvSplit(),
        },
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "target",
                Usage: "rpc target to talk to.",
                Required: true,
            },
        },
    }
}

func cmdCfgMkvInfo() *cli.Command {
    return &cli.Command{
        Name: "info",
        Usage: "get info on MKV file",
        ArgsUsage: "<path>",
        Description: "get info on MKV file.",
        Action: cmdMkvInfo,
    }
}

func cmdMkvInfo(c *cli.Context) error {
    args := c.Args().Slice()
    if len(args) != 1 {
        return errors.New("Expected a single argument")
    }
    path, err := filepath.Abs(args[0])
    if err != nil {
        return fmt.Errorf("Could not determine absolute path name: %w", err)
    }
    target := c.String("target")
    creds := grpc.WithTransportCredentials(insecure.NewCredentials())
    conn, err := grpc.DialContext(c.Context, target, creds)
    if  err != nil {
        return fmt.Errorf("when dialing %w", err)
    }
    defer conn.Close()
    client := muspb.NewMkvUtilClient(conn)
    resp, err := client.GetInfo(c.Context, &muspb.GetInfoRequest{
        InPath: path,
    })
    if err != nil {
        return err
    }    

    _, err = fmt.Fprintf(c.App.Writer, "%s\n", resp)
    return err
}

func cmdCfgMkvSplit() *cli.Command {
    return &cli.Command{
        Name: "split",
        Usage: "split an MKV file into parts",
        ArgsUsage: "-in /path/to/input -out 2-3:/path/to/out1 -out -2:/path/to/out2 -out 4-:/path/to/out3",
        Description: "split an MKV file into parts",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "in",
                Usage: ".mkv file to read",
                Required: true,
            },
            &cli.StringSliceFlag{
                Name: "out",
                Usage: "chapters to export & path to store output at.",
                Required: true,
            },
        },
        Action: cmdMkvSplit,
    }
}

func cmdMkvSplit(c *cli.Context) error {
    in := c.String("in")
    outs := c.StringSlice("out")
    _, err := fmt.Fprintf(c.App.Writer, "in: %s\n", in)
    if err != nil {
        return err
    }
    for _, out := range outs {
        _, err := fmt.Fprintf(c.App.Writer, "out: %s\n", out)
        if err != nil {
            return err
        }
    }
    return nil
}

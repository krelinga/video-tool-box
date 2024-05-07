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

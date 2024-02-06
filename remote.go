package main

import (
    "fmt"
    "errors"

    cli "github.com/urfave/cli/v2"
    "github.com/krelinga/video-tool-box/pb"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func subcmdCfgRemote() *cli.Command {
    return &cli.Command{
        Name: "remote",
        Usage: "Remote operations on video files.",
        Subcommands: []*cli.Command{
            cmdCfgHello(),
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

func cmdCfgHello() *cli.Command {
    return &cli.Command{
        Name: "hello",
        Usage: "hello world to remote transcoding server.",
        Action: cmdHello,
    }
}

func cmdHello(c *cli.Context) error {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return errors.New("toolPaths not present in context")
    }
    args := c.Args().Slice()
    if len(args) != 1 {
        return errors.New("Expected a single file path")
    }
    mountPath := args[0]
    canonPath, err := tp.TranslateNasDir(mountPath)
    if err != nil {
        return err
    }
    conn, err := grpc.DialContext(c.Context, c.String("target"), grpc.WithTransportCredentials(insecure.NewCredentials()))
    if  err != nil {
        return errors.New(fmt.Sprintf("when dialing: %s", err))
    }
    defer conn.Close()

    client := pb.NewTCServerClient(conn)

    resp, err := client.HelloWorld(c.Context, &pb.HelloWorldRequest{In: canonPath})
    if err != nil {
        return errors.New(fmt.Sprintf("from RPC: %s", err))
    }
    fmt.Println(resp)
    return nil
}

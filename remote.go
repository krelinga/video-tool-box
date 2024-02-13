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
            cmdCfgStart(),
            cmdCfgCheck(),
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
        return fmt.Errorf("when dialing %w", err)
    }
    defer conn.Close()

    client := pb.NewTCServerClient(conn)

    resp, err := client.HelloWorld(c.Context, &pb.HelloWorldRequest{In: canonPath})
    if err != nil {
        return fmt.Errorf("from RPC: %w", err)
    }
    fmt.Println(resp)
    return nil
}

func cmdCfgStart() *cli.Command {
    return &cli.Command{
        Name: "start",
        Usage: "start an async transcode on the server.",
        Action: cmdAsyncTranscodeStart,
    }
}

func cmdAsyncTranscodeStart(c *cli.Context) error {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return errors.New("toolPaths not present in context")
    }
    args := c.Args().Slice()
    if len(args) != 3 {
        return errors.New("Expected a name and two file paths")
    }

    name := args[0]
    if len(name) == 0 {
        return errors.New("name must be non-empty")
    }
    inPath, err := tp.TranslateNasDir(args[1])
    if err != nil {
        return err
    }
    outPath, err := tp.TranslateNasDir(args[2])
    if err != nil {
        return err
    }

    conn, err := grpc.DialContext(c.Context, c.String("target"), grpc.WithTransportCredentials(insecure.NewCredentials()))
    if  err != nil {
        return fmt.Errorf("when dialing: %w", err)
    }
    defer conn.Close()

    client := pb.NewTCServerClient(conn)

    _, err = client.StartAsyncTranscode(c.Context, &pb.StartAsyncTranscodeRequest{Name: name, InPath: inPath, OutPath: outPath})

    return err
}

func cmdCfgCheck() *cli.Command {
    return &cli.Command{
        Name: "check",
        Usage: "check on an async transcode on the server.",
        Action: cmdAsyncTranscodeCheck,
    }
}

func cmdAsyncTranscodeCheck(c *cli.Context) error {
    args := c.Args().Slice()
    if len(args) != 1 {
        return errors.New("Expected a name")
    }

    name := args[0]
    if len(name) == 0 {
        return errors.New("name must be non-empty")
    }

    conn, err := grpc.DialContext(c.Context, c.String("target"), grpc.WithTransportCredentials(insecure.NewCredentials()))
    if  err != nil {
        return fmt.Errorf("when dialing: %w", err)
    }
    defer conn.Close()

    client := pb.NewTCServerClient(conn)

    reply, err := client.CheckAsyncTranscode(c.Context, &pb.CheckAsyncTranscodeRequest{Name: name})

    if err != nil {
        return err
    }

    fmt.Fprintln(c.App.Writer, reply)
    if len(reply.ErrorMessage) > 0 {
        fmt.Fprintf(c.App.Writer, "Error Message: %s\n", reply.ErrorMessage)
    }
    return nil
}

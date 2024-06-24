package main

import (
    "fmt"
    "errors"
    "io"
    "os/exec"
    "path/filepath"
    "text/tabwriter"
    "time"

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
            cmdCfgStart(),
            cmdCfgCheck(),
            cmdCfgStartShow(),
            cmdCfgCheckShow(),
            cmdCfgStartSpread(),
            cmdCfgCheckSpread(),
        },
    }
}

func cmdCfgStart() *cli.Command {
    return &cli.Command{
        Name: "start",
        Usage: "start an async transcode on the server.",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "profile",
                Value: "",  // Use the server-side default.
                Usage: "Profile to use for transcoding.",
            },
        },
        Action: cmdAsyncTranscodeStart,
    }
}

func dialTcServer(c *cli.Context) (pb.TCServerClient, func(), error) {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return nil, nil, errors.New("toolPaths not present in context")
    }
    cfg, err := readConfig(tp.ConfigPath())
    if err != nil {
        return nil, nil, err
    }
    creds := grpc.WithTransportCredentials(insecure.NewCredentials())
    conn, err := grpc.DialContext(c.Context, cfg.TcServerTarget, creds)
    if  err != nil {
        return nil, nil, fmt.Errorf("when dialing %w", err)
    }
    cleanup := func() {
        conn.Close()
    }
    client := pb.NewTCServerClient(conn)
    return client, cleanup, nil
}

func cmdAsyncTranscodeStart(c *cli.Context) error {
    args := c.Args().Slice()
    if len(args) != 3 {
        return errors.New("Expected a name and two file paths")
    }

    name := args[0]
    if len(name) == 0 {
        return errors.New("name must be non-empty")
    }
    inPath, err := filepath.Abs(args[1])
    if err != nil {
        return err
    }
    outPath, err := filepath.Abs(args[2])
    if err != nil {
        return err
    }

    client, cleanup, err := dialTcServer(c)
    if err != nil {
        return err
    }
    defer cleanup()

    req := &pb.StartAsyncTranscodeRequest{
        Name: name,
        InPath: inPath,
        OutPath: outPath,
        Profile: c.String("profile"),
    }

    _, err = client.StartAsyncTranscode(c.Context, req)

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

    client, cleanup, err := dialTcServer(c)
    if err != nil {
        return err
    }
    defer cleanup()

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

func cmdCfgStartShow() *cli.Command {
    return &cli.Command{
        Name: "startshow",
        Usage: "start an async transcode of a show on the server.",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "profile",
                Value: "",  // Use the server-side default.
                Usage: "Profile to use for transcoding.",
            },
        },
        Action: cmdAsyncTranscodeStartShow,
    }
}

func cmdAsyncTranscodeStartShow(c *cli.Context) error {
    args := c.Args().Slice()
    if len(args) != 3 {
        return errors.New("Expected a name and two file paths")
    }

    name := args[0]
    if len(name) == 0 {
        return errors.New("name must be non-empty")
    }
    inDirPath, err := filepath.Abs(args[1])
    if err != nil {
        return err
    }
    outParentDirPath, err := filepath.Abs(args[2])
    if err != nil {
        return err
    }

    client, cleanup, err := dialTcServer(c)
    if err != nil {
        return err
    }
    defer cleanup()

    req := &pb.StartAsyncShowTranscodeRequest{
        Name: name,
        InDirPath: inDirPath,
        OutParentDirPath: outParentDirPath,
        Profile: c.String("profile"),
    }

    _, err = client.StartAsyncShowTranscode(c.Context, req)

    return err
}

func cmdCfgCheckShow() *cli.Command {
    return &cli.Command{
        Name: "checkshow",
        Usage: "check on an async show transcode on the server.",
        Action: cmdAsyncTranscodeCheckShow,
    }
}

func cmdAsyncTranscodeCheckShow(c *cli.Context) error {
    args := c.Args().Slice()
    if len(args) != 1 {
        return errors.New("Expected a name")
    }

    name := args[0]
    if len(name) == 0 {
        return errors.New("name must be non-empty")
    }

    client, cleanup, err := dialTcServer(c)
    if err != nil {
        return err
    }
    defer cleanup()

    reply, err := client.CheckAsyncShowTranscode(c.Context, &pb.CheckAsyncShowTranscodeRequest{Name: name})
    if err != nil {
        return err
    }

    fmt.Fprintf(c.App.Writer, "Non-Episode State: %s\n", reply.State)
    if len(reply.ErrorMessage) > 0 {
        fmt.Fprintf(c.App.Writer, "Non-Episode Error Message: %s\n", reply.ErrorMessage)
    }
    fmt.Fprintf(c.App.Writer, "Episodes:\n")
    fmt.Fprintf(c.App.Writer, "=========\n")
    tw := tabwriter.NewWriter(c.App.Writer, 0, 4, 3, byte(' '), 0)
    fmt.Fprintln(tw, "index\tepisode\tstate\tprogress/error")
    fmt.Fprintln(tw, "-----\t-------\t-----\t--------------")
    for i, f := range reply.File {
        progOrError := func() string {
            if len(f.ErrorMessage) > 0 {
                return f.ErrorMessage
            }
            return f.Progress
        }
        fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", i, f.Episode, f.State, progOrError())
    }

    return tw.Flush()
}

func cmdCfgStartSpread() *cli.Command {
    return &cli.Command{
        Name: "startspread",
        Usage: "start an async transcode of a file with multiple profiles on the server.",
        Flags: []cli.Flag{
            &cli.StringSliceFlag{
                Name: "profile",
                Usage: "Profile to use for transcoding.",
                Required: true,
            },
        },
        Action: cmdAsyncTranscodeStartSpread,
    }
}

func cmdAsyncTranscodeStartSpread(c *cli.Context) error {
    args := c.Args().Slice()
    if len(args) != 3 {
        return errors.New("Expected a name and two file paths")
    }

    name := args[0]
    if len(name) == 0 {
        return errors.New("name must be non-empty")
    }
    inPath, err := filepath.Abs(args[1])
    if err != nil {
        return err
    }
    outParentDirPath, err := filepath.Abs(args[2])
    if err != nil {
        return err
    }

    client, cleanup, err := dialTcServer(c)
    if err != nil {
        return err
    }
    defer cleanup()

    req := &pb.StartAsyncSpreadTranscodeRequest{
        Name: name,
        InPath: inPath,
        OutParentDirPath: outParentDirPath,
        ProfileList: &pb.StartAsyncSpreadTranscodeRequest_ProfileList{
            Profile: c.StringSlice("profile"),
        },
    }

    _, err = client.StartAsyncSpreadTranscode(c.Context, req)

    return err
}

var watchFlag = &cli.BoolFlag{
    Name: "watch",
    Usage: "check for progress in a loop.",
}

func clearScreen(out io.Writer) error {
    cmd := exec.Command("clear")
    cmd.Stdout = out
    return cmd.Run()
}

func cmdCfgCheckSpread() *cli.Command {
    return &cli.Command{
        Name: "checkspread",
        Usage: "check on an async spread transcode on the server.",
        Flags: []cli.Flag{
            watchFlag,
        },
        Action: cmdAsyncTranscodeCheckSpread,
    }
}

func cmdAsyncTranscodeCheckSpread(c *cli.Context) error {
    args := c.Args().Slice()
    if len(args) != 1 {
        return errors.New("Expected a name")
    }

    name := args[0]
    if len(name) == 0 {
        return errors.New("name must be non-empty")
    }

    client, cleanup, err := dialTcServer(c)
    if err != nil {
        return err
    }
    defer cleanup()

    for {
        reply, err := client.CheckAsyncSpreadTranscode(c.Context, &pb.CheckAsyncSpreadTranscodeRequest{Name: name})
        if err != nil {
            return err
        }
        if c.Bool("watch") {
            if err := clearScreen(c.App.Writer); err != nil {
                return err
            }
        }
        fmt.Fprintf(c.App.Writer, "State: %s\n", reply.State)
        if len(reply.ErrorMessage) > 0 {
            fmt.Fprintf(c.App.Writer, "Error Message: %s\n", reply.ErrorMessage)
        }
        fmt.Fprintf(c.App.Writer, "Profiles:\n")
        fmt.Fprintf(c.App.Writer, "=========\n")
        tw := tabwriter.NewWriter(c.App.Writer, 0, 4, 3, byte(' '), 0)
        fmt.Fprintln(tw, "index\tprofile\tstate\tprogress/error")
        fmt.Fprintln(tw, "-----\t-------\t-----\t--------------")
        for i, f := range reply.Profile {
            progOrError := func() string {
                if len(f.ErrorMessage) > 0 {
                    return f.ErrorMessage
                }
                return f.Progress
            }
            fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", i, f.Profile, f.State, progOrError())
        }
        if err := tw.Flush(); err != nil {
            return err
        }

        if !c.Bool("watch") {
            break
        }
        time.Sleep(time.Second * 5)
    }

    return nil
}

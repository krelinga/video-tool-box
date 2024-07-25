package main

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "os/exec"
    "path/filepath"
    "strings"
    "text/tabwriter"
    "time"

    "connectrpc.com/connect"
    cli "github.com/urfave/cli/v2"

    pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/tcserver/v1"
    pbconnect "buf.build/gen/go/krelinga/proto/connectrpc/go/krelinga/video/tcserver/v1/tcserverv1connect"
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
            cmdCfgRemoteList(),
            cmdCfgStartMovie(),
            cmdCfgCheckMovie(),
        },
    }
}

var watchFlag = &cli.BoolFlag{
    Name: "watch",
    Usage: "check for progress in a loop.",
}

var requiredProfileFlag = &cli.StringFlag{
    Name: "profile",
    Value: "",
    Usage: "Profile to use for transcoding.",
    Required: true,
}

var nameFlag = &cli.StringFlag{
    Name: "name",
    Value: "",
    Usage: "Name to use on the transcoding server.",
}

var requiredNameFlag = func() *cli.StringFlag {
    var f = *nameFlag
    f.Required = true
    return &f
}()

func requiredInFlag(usage string) *cli.StringFlag {
    f := &cli.StringFlag{
        Name: "in",
        Required: true,
        Usage: usage,
    }
    return f
}

func outFlag(usage string) *cli.StringFlag {
    f := &cli.StringFlag{
        Name: "out",
        Usage: usage,
    }
    return f
}

func requiredOutFlag(usage string) *cli.StringFlag {
    f := outFlag(usage)
    f.Required = true
    return f
}

func clearScreen(out io.Writer) error {
    cmd := exec.Command("clear")
    cmd.Stdout = out
    return cmd.Run()
}

func formatTranscodeState(s pb.TranscodeState) string {
    str := s.String()
    return strings.TrimPrefix(str, "TRANSCODE_STATE_")
}

func cmdCfgStart() *cli.Command {
    return &cli.Command{
        Name: "start",
        Usage: "start an async transcode on the server.",
        Flags: []cli.Flag{
            requiredProfileFlag,
            nameFlag,
            requiredInFlag("Path ending in .mkv to read"),
            requiredOutFlag("Path ending in .mkv to write"),
        },
        Action: cmdAsyncTranscodeStart,
    }
}

func remoteInit(c *cli.Context) (pbconnect.TCServiceClient, *config, error) {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return nil, nil, errors.New("toolPaths not present in context")
    }
    cfg, err := readConfig(tp.ConfigPath())
    if err != nil {
        return nil, nil, err
    }
    return pbconnect.NewTCServiceClient(http.DefaultClient, cfg.TcServerTarget), cfg, nil
}

func cmdAsyncTranscodeStart(c *cli.Context) error {
    inPath, err := filepath.Abs(c.String("in"))
    if err != nil {
        return err
    }
    outPath, err := filepath.Abs(c.String("out"))
    if err != nil {
        return err
    }

    name := func() string {
        if f := c.String("name"); len(f) > 0 {
            return f
        }
        return filepath.Base(inPath)
    }()

    client, _, err := remoteInit(c)
    if err != nil {
        return err
    }

    req := connect.NewRequest(&pb.StartAsyncTranscodeRequest{
        Name: name,
        InPath: inPath,
        OutPath: outPath,
        Profile: c.String("profile"),
    })

    _, err = client.StartAsyncTranscode(c.Context, req)

    return err
}

func cmdCfgCheck() *cli.Command {
    return &cli.Command{
        Name: "check",
        Usage: "check on an async transcode on the server.",
        Flags: []cli.Flag{
            watchFlag,
            requiredNameFlag,
        },
        Action: cmdAsyncTranscodeCheck,
    }
}

func cmdAsyncTranscodeCheck(c *cli.Context) error {
    name := c.String("name")
    watch := c.Bool("watch")

    client, _, err := remoteInit(c)
    if err != nil {
        return err
    }

    for {
        request := connect.NewRequest(&pb.CheckAsyncTranscodeRequest{Name: name})
        reply, err := client.CheckAsyncTranscode(c.Context, request)
        if err != nil {
            return err
        }
        if watch {
            if err := clearScreen(c.App.Writer); err != nil {
                return err
            }
        }

        fmt.Fprintln(c.App.Writer, reply.Msg)
        if len(reply.Msg.ErrorMessage) > 0 {
            fmt.Fprintf(c.App.Writer, "Error Message: %s\n", reply.Msg.ErrorMessage)
        }
        if !watch {
            break
        }
        time.Sleep(time.Second * 5)
    }
    return nil
}

func cmdCfgStartShow() *cli.Command {
    return &cli.Command{
        Name: "startshow",
        Usage: "start an async transcode of a show on the server.",
        Flags: []cli.Flag{
            requiredProfileFlag,
            requiredInFlag("Directory containing the show to read."),
            nameFlag,
            outFlag("If set, overrides the output directory for all shows"),
        },
        Action: cmdAsyncTranscodeStartShow,
    }
}

func cmdAsyncTranscodeStartShow(c *cli.Context) error {
    inDirPath, err := filepath.Abs(c.String("in"))
    if err != nil {
        return err
    }

    name := func() string {
        if f := c.String("name"); len(f) > 0 {
            return f
        }
        return filepath.Base(inDirPath)
    }()
    fmt.Println(name)

    client, cfg, err := remoteInit(c)
    if err != nil {
        return err
    }

    rawOutDir := func() string {
        if f := c.String("out"); len(f) > 0 {
            return f
        } else {
            return cfg.DefaultShowTranscodeOutDir
        }
    }()
    outDir, err := filepath.Abs(rawOutDir)
    if err != nil {
        return err
    }

    req := connect.NewRequest(&pb.StartAsyncShowTranscodeRequest{
        Name: name,
        InDirPath: inDirPath,
        OutParentDirPath: outDir,
        Profile: c.String("profile"),
    })

    _, err = client.StartAsyncShowTranscode(c.Context, req)

    return err
}

func cmdCfgCheckShow() *cli.Command {
    return &cli.Command{
        Name: "checkshow",
        Usage: "check on an async show transcode on the server.",
        Flags: []cli.Flag{
            watchFlag,
            requiredNameFlag,
        },
        Action: cmdAsyncTranscodeCheckShow,
    }
}

func cmdAsyncTranscodeCheckShow(c *cli.Context) error {
    name := c.String("name")
    watch := c.Bool("watch")

    client, _, err := remoteInit(c)
    if err != nil {
        return err
    }

    for {
        request := connect.NewRequest(&pb.CheckAsyncShowTranscodeRequest{Name: name})
        reply, err := client.CheckAsyncShowTranscode(c.Context, request)
        if err != nil {
            return err
        }

        if watch {
            if err := clearScreen(c.App.Writer); err != nil {
                return err
            }
        }

        fmt.Fprintf(c.App.Writer, "Non-Episode State: %s\n", reply.Msg.State)
        fmt.Fprintf(c.App.Writer, "Profile: %s\n", reply.Msg.Profile)
        if len(reply.Msg.ErrorMessage) > 0 {
            fmt.Fprintf(c.App.Writer, "Non-Episode Error Message: %s\n", reply.Msg.ErrorMessage)
        }
        fmt.Fprintf(c.App.Writer, "Episodes:\n")
        fmt.Fprintf(c.App.Writer, "=========\n")
        tw := tabwriter.NewWriter(c.App.Writer, 0, 4, 3, byte(' '), 0)
        fmt.Fprintln(tw, "index\tepisode\tstate\tprogress/error")
        fmt.Fprintln(tw, "-----\t-------\t-----\t--------------")
        for i, f := range reply.Msg.File {
            progOrError := func() string {
                if len(f.ErrorMessage) > 0 {
                    return f.ErrorMessage
                }
                return f.Progress
            }
            fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", i, f.Episode, formatTranscodeState(f.State), progOrError())
        }
        if err := tw.Flush(); err != nil {
            return err
        }
        if watch {
            break
        }
        time.Sleep(time.Second * 5)
    }

    return nil
}

func cmdCfgStartSpread() *cli.Command {
    return &cli.Command{
        Name: "startspread",
        Usage: "start an async transcode of a file with multiple profiles on the server.",
        Flags: []cli.Flag{
            requiredProfileFlag,
            requiredNameFlag,
            requiredInFlag("A path ending in .mkv to read."),
            requiredOutFlag("A directory that will be created to store spread output."),
        },
        Action: cmdAsyncTranscodeStartSpread,
    }
}

func cmdAsyncTranscodeStartSpread(c *cli.Context) error {
    name := c.String("name")
    inPath, err := filepath.Abs(c.String("in"))
    if err != nil {
        return err
    }
    outParentDirPath, err := filepath.Abs(c.String("out"))
    if err != nil {
        return err
    }

    client, _, err := remoteInit(c)
    if err != nil {
        return err
    }

    req := connect.NewRequest(&pb.StartAsyncSpreadTranscodeRequest{
        Name: name,
        InPath: inPath,
        OutParentDirPath: outParentDirPath,
        ProfileList: &pb.StartAsyncSpreadTranscodeRequest_ProfileList{
            Profile: c.StringSlice("profile"),
        },
    })

    _, err = client.StartAsyncSpreadTranscode(c.Context, req)

    return err
}

func cmdCfgCheckSpread() *cli.Command {
    return &cli.Command{
        Name: "checkspread",
        Usage: "check on an async spread transcode on the server.",
        Flags: []cli.Flag{
            watchFlag,
            requiredNameFlag,
        },
        Action: cmdAsyncTranscodeCheckSpread,
    }
}

func cmdAsyncTranscodeCheckSpread(c *cli.Context) error {
    name := c.String("name")
    watch := c.Bool("watch")

    client, _, err := remoteInit(c)
    if err != nil {
        return err
    }

    for {
        request := connect.NewRequest(&pb.CheckAsyncSpreadTranscodeRequest{Name: name})
        reply, err := client.CheckAsyncSpreadTranscode(c.Context, request)
        if err != nil {
            return err
        }
        if watch {
            if err := clearScreen(c.App.Writer); err != nil {
                return err
            }
        }
        fmt.Fprintf(c.App.Writer, "State: %s\n", reply.Msg.State)
        if len(reply.Msg.ErrorMessage) > 0 {
            fmt.Fprintf(c.App.Writer, "Error Message: %s\n", reply.Msg.ErrorMessage)
        }
        fmt.Fprintf(c.App.Writer, "Profiles:\n")
        fmt.Fprintf(c.App.Writer, "=========\n")
        tw := tabwriter.NewWriter(c.App.Writer, 0, 4, 3, byte(' '), 0)
        fmt.Fprintln(tw, "index\tprofile\tstate\tprogress/error")
        fmt.Fprintln(tw, "-----\t-------\t-----\t--------------")
        for i, f := range reply.Msg.Profile {
            progOrError := func() string {
                if len(f.ErrorMessage) > 0 {
                    return f.ErrorMessage
                }
                return f.Progress
            }
            fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", i, f.Profile, formatTranscodeState(f.State), progOrError())
        }
        if err := tw.Flush(); err != nil {
            return err
        }

        if !watch {
            break
        }
        time.Sleep(time.Second * 5)
    }

    return nil
}

func cmdCfgRemoteList() *cli.Command {
    return &cli.Command{
        Name: "list",
        Usage: "List async transcode operations of all types",
        Action: cmdRemoteList,
    }
}

func cmdRemoteList(c *cli.Context) error {
    client, _, err := remoteInit(c)
    if err != nil {
        return err
    }
    request := connect.NewRequest(&pb.ListAsyncTranscodesRequest{})
    reply, err := client.ListAsyncTranscodes(c.Context, request)
    if err != nil {
        return err
    }
    tw := tabwriter.NewWriter(c.App.Writer, 0, 4, 3, byte(' '), 0)
    fmt.Fprintln(tw, "name\ttype\tstate")
    fmt.Fprintln(tw, "----\t----\t-----")
    for _, op := range reply.Msg.Op {
        fmt.Fprintf(tw, "%s\t%s\t%s\n", op.Name, op.Type, op.State)
    }
    return tw.Flush()
}

func cmdCfgStartMovie() *cli.Command {
    return &cli.Command{
        Name: "startmovie",
        Usage: "start an async transcode of a movie on the server.",
        Flags: []cli.Flag{
            requiredProfileFlag,
            nameFlag,
            requiredInFlag("Directory containing the movie to be read."),
            outFlag("If set, overrides the default output directory for all movies."),
        },
        Action: cmdAsyncTranscodeStartMovie,
    }
}

func cmdAsyncTranscodeStartMovie(c *cli.Context) error {
    inDirPath, err := filepath.Abs(c.String("in"))
    if err != nil {
        return err
    }
    in := filepath.Join(inDirPath, filepath.Base(inDirPath) + ".mkv")

    name := func() string {
        if f := c.String("name"); len(f) > 0 {
            return f
        }
        return filepath.Base(inDirPath)
    }()
    fmt.Println(name)

    client, cfg, err := remoteInit(c)
    if err != nil {
        return err
    }

    rawOutDir := func() string {
        if f := c.String("out"); len(f) > 0 {
            return f
        } else {
            return cfg.DefaultMovieTranscodeOutDir
        }
    }()
    outDir, err := filepath.Abs(rawOutDir)
    if err != nil {
        return err
    }
    out := filepath.Join(outDir, filepath.Base(inDirPath), filepath.Base(inDirPath) + ".mkv")

    req := connect.NewRequest(&pb.StartAsyncTranscodeRequest{
        Name: name,
        InPath: in,
        OutPath: out,
        Profile: c.String("profile"),
    })

    _, err = client.StartAsyncTranscode(c.Context, req)

    return err
}

func cmdCfgCheckMovie() *cli.Command {
    return &cli.Command{
        Name: "checkmovie",
        Usage: "check on an async movie transcode on the server.",
        Flags: []cli.Flag{
            watchFlag,
            requiredNameFlag,
        },
        Action: cmdAsyncTranscodeCheckMovie,
    }
}

func cmdAsyncTranscodeCheckMovie(c *cli.Context) error {
    name := c.String("name")
    watch := c.Bool("watch")

    client, _, err := remoteInit(c)
    if err != nil {
        return err
    }

    for {
        request := connect.NewRequest(&pb.CheckAsyncTranscodeRequest{Name: name})
        reply, err := client.CheckAsyncTranscode(c.Context, request)
        if err != nil {
            return err
        }

        if watch {
            if err := clearScreen(c.App.Writer); err != nil {
                return err
            }
        }

        fmt.Fprintf(c.App.Writer, "Name: %s\n", name)
        fmt.Fprintf(c.App.Writer, "State: %s\n", reply.Msg.State)
        fmt.Fprintf(c.App.Writer, "Profile: %s\n", reply.Msg.Profile)
        if len(reply.Msg.ErrorMessage) > 0 {
            fmt.Fprintf(c.App.Writer, "Error Message: %s\n", reply.Msg.ErrorMessage)
        }
        if len(reply.Msg.Progress) > 0 {
            fmt.Fprintf(c.App.Writer, "Progress: %s\n", reply.Msg.Progress)
        }
        if !watch {
            break
        }
        time.Sleep(time.Second * 5)
    }

    return nil
}

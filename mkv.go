package main

import (
    "errors"
    "fmt"
    "path/filepath"
    "regexp"
    "strconv"
    "text/tabwriter"

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
            cmdCfgMkvChapters(),
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

var splitSpecRe = regexp.MustCompile(`(\d+)?-(\d+)?:(.+)`)

func cmdMkvSplit(c *cli.Context) error {
    in, err := filepath.Abs(c.String("in"))
    if err != nil {
        return fmt.Errorf("Could not get absolute path: %w", err)
    }
    outs := []*muspb.SplitRequest_ByChapters{}
    for _, o := range c.StringSlice("out") {
        match := splitSpecRe.FindStringSubmatch(o)
        if match == nil {
            return fmt.Errorf("Could not parse --out %s", o)
        }
        c := &muspb.SplitRequest_ByChapters{}
        var err error
        c.OutPath, err = filepath.Abs(match[3])
        if err != nil {
            return fmt.Errorf("Could not get absolute path for --out %s", o)
        }
        atoi := func(s string) (int32, error) {
            i, err := strconv.Atoi(s)
            if err != nil {
                return 0, err
            }
            return int32(i), nil
        }
        if len(match[1]) > 0 {
            c.Start, err = atoi(match[1])
            if err != nil {
                return fmt.Errorf("Could not parse --out %s", o)
            }
        }
        if len(match[2]) > 0 {
            c.Limit, err = atoi(match[2])
            if err != nil {
                return fmt.Errorf("Could not parse --out %s", o)
            }
        }
        outs = append(outs, c)
    }
    req := &muspb.SplitRequest{
        InPath: in,
        ByChapters: outs,
    }
    target := c.String("target")
    creds := grpc.WithTransportCredentials(insecure.NewCredentials())
    conn, err := grpc.DialContext(c.Context, target, creds)
    if  err != nil {
        return fmt.Errorf("when dialing %w", err)
    }
    defer conn.Close()
    client := muspb.NewMkvUtilClient(conn)
    _, err = client.Split(c.Context, req)
    return err
}

func cmdCfgMkvChapters() *cli.Command {
    return &cli.Command{
        Name: "chapters",
        Usage: "get chapters present in an mkv file",
        ArgsUsage: "<path>",
        Description: "get chapters present in an mkv file",
        Action: cmdMkvChapters,
    }
}

func cmdMkvChapters(c *cli.Context) error {
    args := c.Args().Slice()
    if len(args) != 1 {
        return errors.New("Expected a single argument")
    }
    in, err := filepath.Abs(args[0])
    if err != nil {
        return fmt.Errorf("Could not get absolute path: %w", err)
    }
    req := &muspb.GetChaptersRequest{
        InPath: in,
        Format: muspb.ChaptersFormat_CF_SIMPLE,
    }
    target := c.String("target")
    creds := grpc.WithTransportCredentials(insecure.NewCredentials())
    conn, err := grpc.DialContext(c.Context, target, creds)
    if err != nil {
        return fmt.Errorf("when dialing %w", err)
    }
    defer conn.Close()
    client := muspb.NewMkvUtilClient(conn)
    resp, err := client.GetChapters(c.Context, req)
    if err != nil {
        return err
    }

    tw := tabwriter.NewWriter(c.App.Writer, 0, 4, 2, byte(' '), 0)
    _, err = fmt.Fprintf(tw, "Number\tTitle\tOffset\n")
    if err != nil {
        return err
    }
    _, err = fmt.Fprintf(tw, "======\t=====\t======\n")
    if err != nil {
        return err
    }
    for _, c := range resp.Chapters.Simple.Chapters {
        _, err := fmt.Fprintf(tw, "%d\t%s\t%s\n", c.Number, c.Name, c.Offset.AsDuration())
        if err != nil {
            return err
        }
    }
    return tw.Flush()
}

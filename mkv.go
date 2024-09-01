package main

// spell-checker:ignore urfave, connectrpc, muspbconnect, muspp, serverv1, protocolbuffers, muspb, subcmd

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"text/tabwriter"

	"connectrpc.com/connect"
	cli "github.com/urfave/cli/v2"

	muspbconnect "buf.build/gen/go/krelinga/proto/connectrpc/go/krelinga/video/mkv_util_server/v1/mkv_util_serverv1connect"
	muspb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/mkv_util_server/v1"
)

func subcmdCfgMkv() *cli.Command {
	return &cli.Command{
		Name:  "mkv",
		Usage: "Manipulate MKV files",
		Subcommands: []*cli.Command{
			cmdCfgMkvInfo(),
			cmdCfgMkvSplit(),
			cmdCfgMkvConcat(),
			cmdCfgMkvChapters(),
		},
	}
}

// Returns the client, a function to call to clean up the client, and any error.
func dialMkvUtilsServer(c *cli.Context) (muspbconnect.MkvUtilServiceClient, error) {
	tp, ok := toolPathsFromContext(c.Context)
	if !ok {
		return nil, errors.New("toolPaths not present in context")
	}
	cfg, err := readConfig(tp.ConfigPath())
	if err != nil {
		return nil, err
	}
	return muspbconnect.NewMkvUtilServiceClient(http.DefaultClient, cfg.MkvUtilServerTarget), nil
}

func cmdCfgMkvInfo() *cli.Command {
	return &cli.Command{
		Name:        "info",
		Usage:       "get info on MKV file",
		ArgsUsage:   "<path>",
		Description: "get info on MKV file.",
		Action:      cmdMkvInfo,
	}
}

func cmdMkvInfo(c *cli.Context) error {
	args := c.Args().Slice()
	if len(args) != 1 {
		return errors.New("expected a single argument")
	}
	path, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("could not determine absolute path name: %w", err)
	}
	client, err := dialMkvUtilsServer(c)
	if err != nil {
		return err
	}
	resp, err := client.GetInfo(c.Context, connect.NewRequest(&muspb.GetInfoRequest{
		InPath: path,
	}))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.App.Writer, "%s\n", resp.Msg)
	return err
}

func cmdCfgMkvSplit() *cli.Command {
	return &cli.Command{
		Name:        "split",
		Usage:       "split an MKV file into parts",
		ArgsUsage:   "-in /path/to/input -out 2-3:/path/to/out1 -out -2:/path/to/out2 -out 4-:/path/to/out3",
		Description: "split an MKV file into parts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "in",
				Usage:    ".mkv file to read",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "out",
				Usage:    "chapters to export & path to store output at.",
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
		return fmt.Errorf("could not get absolute path: %w", err)
	}
	outs := []*muspb.SplitRequest_ByChapters{}
	for _, o := range c.StringSlice("out") {
		match := splitSpecRe.FindStringSubmatch(o)
		if match == nil {
			return fmt.Errorf("could not parse --out %s", o)
		}
		c := &muspb.SplitRequest_ByChapters{}
		var err error
		c.OutPath, err = filepath.Abs(match[3])
		if err != nil {
			return fmt.Errorf("could not get absolute path for --out %s", o)
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
				return fmt.Errorf("could not parse --out %s", o)
			}
		}
		if len(match[2]) > 0 {
			c.Limit, err = atoi(match[2])
			if err != nil {
				return fmt.Errorf("could not parse --out %s", o)
			}
		}
		outs = append(outs, c)
	}
	req := connect.NewRequest(&muspb.SplitRequest{
		InPath:     in,
		ByChapters: outs,
	})
	client, err := dialMkvUtilsServer(c)
	if err != nil {
		return err
	}
	_, err = client.Split(c.Context, req)
	return err
}

func cmdCfgMkvChapters() *cli.Command {
	return &cli.Command{
		Name:        "chapters",
		Usage:       "get chapters present in an mkv file",
		ArgsUsage:   "<path>",
		Description: "get chapters present in an mkv file",
		Action:      cmdMkvChapters,
	}
}

func cmdMkvChapters(c *cli.Context) error {
	args := c.Args().Slice()
	if len(args) != 1 {
		return errors.New("expected a single argument")
	}
	in, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("could not get absolute path: %w", err)
	}
	req := connect.NewRequest(&muspb.GetChaptersRequest{
		InPath: in,
		Format: muspb.ChaptersFormat_CHAPTERS_FORMAT_SIMPLE,
	})
	client, err := dialMkvUtilsServer(c)
	if err != nil {
		return err
	}
	resp, err := client.GetChapters(c.Context, req)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(c.App.Writer, 0, 4, 2, byte(' '), 0)
	_, err = fmt.Fprintf(tw, "Number\tTitle\tOffset\tDuration\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(tw, "======\t=====\t======\t========\n")
	if err != nil {
		return err
	}
	for _, c := range resp.Msg.Chapters.Simple.Chapters {
		_, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", c.Number, c.Name, c.Offset.AsDuration(), c.Duration.AsDuration())
		if err != nil {
			return err
		}
	}
	return tw.Flush()
}

func cmdCfgMkvConcat() *cli.Command {
	return &cli.Command{
		Name:        "concat",
		Usage:       "concatenate MKV files into a larger file.",
		ArgsUsage:   "-in /path/to/input1 -in path/to/input2 -out /path/to/out",
		Description: "split an MKV file into parts",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "in",
				Usage:    ".mkv file to read",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "out",
				Usage:    "path to write combined file at",
				Required: true,
			},
		},
		Action: cmdMkvConcat,
	}
}

func cmdMkvConcat(c *cli.Context) error {
	req := connect.NewRequest(&muspb.ConcatRequest{})
	for _, in := range c.StringSlice("in") {
		fullPath, err := filepath.Abs(in)
		if err != nil {
			return fmt.Errorf("could not get absolute path: %w", err)
		}
		req.Msg.InputPaths = append(req.Msg.InputPaths, fullPath)
	}
	fullPath, err := filepath.Abs(c.String("out"))
	if err != nil {
		return fmt.Errorf("could not get absolute path: %w", err)
	}
	req.Msg.OutputPath = fullPath

	client, err := dialMkvUtilsServer(c)
	if err != nil {
		return err
	}

	_, err = client.Concat(c.Context, req)
	return err
}

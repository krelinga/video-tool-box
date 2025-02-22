package main

// spell-checker:ignore urfave, subcmd, Tcprofiles

import cli "github.com/urfave/cli/v2"

func appCfg() *cli.App {
	return &cli.App{
		Name:  "vtb",
		Usage: "Video Tool Box",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "handbrake",
				Usage: "path to handbrake binary for transcoding.",
			},
		},
		Commands: []*cli.Command{
			subcmdCfgRip(),
			subcmdCfgRemote(),
			subcmdCfgMkv(),
			cmdCfgConfig(),
			subcmdCfgCache(),
			cmdCfgWriteTcprofiles(),
		},
		// Caller should set these.
		Reader:    nil,
		Writer:    nil,
		ErrWriter: nil,
	}
}

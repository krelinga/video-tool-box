package main

// spell-checker:ignore urfave subcmd subdirs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
)

func subcmdCfgCache() *cli.Command {
	return &cli.Command{
		Name:  "cache",
		Usage: "Manipulate the server-side rip cache.",
		Subcommands: []*cli.Command{
			cmdCfgSync(),
			cmdCfgClear(),
		},
	}
}

func cacheInit(c *cli.Context) (*config, error) {
	tp, ok := toolPathsFromContext(c.Context)
	if !ok {
		return nil, errors.New("toolPaths not present in context")
	}
	cfg, err := readConfig(tp.ConfigPath())
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func cmdCfgSync() *cli.Command {
	return &cli.Command{
		Name:   "sync",
		Usage:  "Sync server-side cache to laptop.",
		Action: cmdSync,
	}
}

func cmdSync(c *cli.Context) error {
	cfg, err := cacheInit(c)
	if err != nil {
		return err
	}

	subdirs, err := filepath.Glob(filepath.Join(cfg.RipCacheServerDir, "*"))
	if err != nil {
		return err
	}
	args := []string{
		"-ah",
		"--progress",
		"-r",
	}
	args = append(args, subdirs...)
	args = append(args, cfg.RipCacheLocalDir)
	cmd := exec.Command("/usr/bin/rsync", args...)
	cmd.Stdin = c.App.Reader
	cmd.Stdout = c.App.Writer
	cmd.Stderr = c.App.ErrWriter
	return cmd.Run()
}

func cmdCfgClear() *cli.Command {
	return &cli.Command{
		Name:   "clear",
		Usage:  "Clear server-side cache.",
		Action: cmdClear,
	}
}

func cmdClear(c *cli.Context) error {
	cfg, err := cacheInit(c)
	if err != nil {
		return err
	}

	subdirs, err := filepath.Glob(filepath.Join(cfg.RipCacheServerDir, "*"))
	if err != nil {
		return err
	}

	fmt.Fprintln(c.App.Writer, "Will delete the following:")
	for _, sd := range subdirs {
		fmt.Fprintf(c.App.Writer, "- %s\n", sd)
	}
	fmt.Fprintf(c.App.Writer, "Confirm (y/N)? ")
	var confirm string
	fmt.Fscanf(c.App.Reader, "%s", &confirm)
	if confirm != "y" {
		return nil
	}

	fmt.Fprintf(c.App.Writer, "Deleting... ")
	for _, sd := range subdirs {
		if err := os.RemoveAll(sd); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.App.Writer, "Done\n")

	return nil
}

package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"

	humanize "github.com/dustin/go-humanize"
	uuid "github.com/google/uuid"
	cli "github.com/urfave/cli/v2"
)

func listMkvFilePaths(currentDir string) ([]string, error) {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		path := filepath.Join(currentDir, entry.Name())
		if filepath.Ext(path) != ".mkv" {
			continue
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func openInVLC(path string) error {
	cmd := exec.Command("open", "-a", "/Applications/VLC.app", path)
	return cmd.Run()
}

func createDestDirAndMove(toMove string, destDir string) error {
	exists := func(path string) error {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("path %s already exists", path)
	}
	basename := filepath.Base(toMove)
	destPath := filepath.Join(destDir, uuid.NewString()+"-"+basename)
	if err := exists(destPath); err != nil {
		return err
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	return os.Rename(toMove, destPath)
}

func readableFileSize(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	bigSize := big.NewInt(info.Size())
	return humanize.BigIBytes(bigSize), nil
}

func cmdCfgDir() *cli.Command {
	return &cli.Command{
		Name:        "dir",
		Usage:       "process .mkv files in a directory one at a time",
		Description: "Requires an existing project.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Name of the project to add files to.",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "dir",
				Usage: "The directory to process.  PWD by default.",
			},
		},
		Action: cmdDir,
	}
}

func cmdDir(c *cli.Context) error {
	tp, ts, _, err := ripCmdInit(c)
	if err != nil {
		return err
	}

	name := c.String("name")
	project, found := ts.FindByName(name)
	if !found {
		return fmt.Errorf("no project named %s", name)
	}

	rootDir, err := func() (string, error) {
		if f := c.String("dir"); len(f) > 0 {
			return filepath.Abs(f)
		}
		return tp.CurrentDir(), nil
	}()
	if err != nil {
		return err
	}

	paths, err := listMkvFilePaths(rootDir)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(c.App.Reader)
	prompt := func() (string, error) {
		fmt.Fprintf(c.App.Writer, "(o)pen, (t)itle, e(x)tra, (s)kip, (d)elete, (q)uit: ")
		if !scanner.Scan() {
			return "", scanner.Err()
		}
		return scanner.Text(), nil
	}
	printPath := func(path string) error {
		size, err := readableFileSize(path)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.App.Writer, "\n%s: %s\n", filepath.Base(path), size)
		return nil
	}

pathLoop:
	for _, path := range paths {
		if err := printPath(path); err != nil {
			return err
		}
	inputLoop:
		for {
			in, err := prompt()
			if err != nil {
				return err
			}
			switch in {
			case "o":
				if err := openInVLC(path); err != nil {
					return err
				}
				fmt.Fprintln(c.App.Writer, "opened in VLC")
				// Repeat inputLoop
			case "t":
				destDir, err := tp.TmmProjectDir(project)
				if err != nil {
					return err
				}
				if err := createDestDirAndMove(path, destDir); err != nil {
					return err
				}
				fmt.Fprintln(c.App.Writer, "moved to TMM content dir")
				break inputLoop
			case "x":
				destDir, err := tp.TmmProjectExtrasDir(project)
				if err != nil {
					return err
				}
				if err := createDestDirAndMove(path, destDir); err != nil {
					return err
				}
				fmt.Fprintln(c.App.Writer, "moved to extras dir")
				break inputLoop
			case "s":
				fmt.Fprintln(c.App.Writer, "skipped")
				continue pathLoop
			case "d":
				if err := os.Remove(path); err != nil {
					return err
				}
				fmt.Fprintln(c.App.Writer, "deleted")
				break inputLoop
			case "q":
				fmt.Fprintln(c.App.Writer, "quit")
				break pathLoop
			}
		}
	}

	ripDirEmpty := func() bool {
		entries, err := os.ReadDir(rootDir)
		if err != nil {
			// Just swallow it ... this is only an optimization.
			return false
		}
		return len(entries) == 0
	}()
	if rootDir != tp.CurrentDir() && ripDirEmpty {
		fmt.Fprintf(c.App.Writer, "rip dir %s empty, delete it (y/N)? ", rootDir)
		var confirm string
		fmt.Fscanf(c.App.Reader, "%s", &confirm)
		if confirm != "y" {
			return nil
		}
		return os.Remove(rootDir)
	}

	return nil
}

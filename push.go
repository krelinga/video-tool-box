package main

// spellchecker: ignore urfave

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	humanize "github.com/dustin/go-humanize"
	mac "github.com/krelinga/go-lib/mac"
	cli "github.com/urfave/cli/v2"
)

func cmdCfgPush() *cli.Command {
	return &cli.Command{
		Name:  "push",
		Usage: "push files from Tiny Media Manager directory to NAS.",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "name",
				Usage: "If set, only named projects will be pushed.",
			},
		},
		Action: cmdPush,
	}
}

func dirBytes(path string) (int64, error) {
	total := int64(0)
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	}
	if err := fs.WalkDir(os.DirFS(path), ".", walkFn); err != nil {
		return 0, err
	}
	return total, nil
}

func cmdPush(c *cli.Context) error {
	tp, ts, save, err := ripCmdInit(c)
	if err != nil {
		return err
	}

	// Prevent macos from sleeping during the push.
	resume, err := mac.StayAwake(mac.StayAwakeOpts{
		System: true,
		Disk:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to prevent sleep: %w", err)
	}
	defer func() {
		if err := resume(); err != nil {
			fmt.Fprintf(c.App.ErrWriter, "failed to resume normal sleep behavior: %v\n", err)
		}
	}()

	projects := ts.FindByStage(psReadyForPush)
	if len(projects) == 0 {
		return errors.New("no projects ready for push")
	}
	if nameList := c.StringSlice("name"); len(nameList) > 0 {
		type empty struct{}
		nameSet := make(map[string]empty)
		for _, n := range nameList {
			nameSet[n] = empty{}
		}

		newProjects := make([]*projectState, 0, len(projects))
		for _, p := range projects {
			if _, found := nameSet[p.Name]; found {
				newProjects = append(newProjects, p)
			}
		}
		projects = newProjects
	}

	dirs := make([]string, 0, len(projects))
	outSuperDirs := make([]string, 0, len(projects))
	dirSizes := make([]int64, 0, len(projects))
	outDirs := make([]string, 0, len(projects))
	totalSize := int64(0)
	for _, p := range projects {
		d, err := tp.TmmProjectDir(p)
		if err != nil {
			return err
		}
		dirs = append(dirs, d)
		size, err := dirBytes(d)
		if err != nil {
			return err
		}
		totalSize += size
		dirSizes = append(dirSizes, size)

		nasSubDir, err := func() (string, error) {
			switch p.Pt {
			case ptUndef:
				return "", errors.New("not in a rip project")
			case ptMovie:
				return "Movies", nil
			case ptShow:
				return "Shows", nil
			default:
				return "", fmt.Errorf("unexpected ProjectType value %v", p.Pt)
			}
		}()
		if err != nil {
			return err
		}
		title := filepath.Base(d)
		outSuperDir := filepath.Join(tp.NasMountDir(), nasSubDir)
		outSuperDirs = append(outSuperDirs, outSuperDir)
		outDir := filepath.Join(outSuperDir, title)
		outDirs = append(outDirs, outDir)
	}

	fmt.Fprintf(c.App.Writer, "Will publish %d projects to NAS as follows:\n", len(dirs))
	for i, dir := range dirs {
		outDir := outDirs[i]
		humanSize := humanize.IBytes(uint64(dirSizes[i]))
		fmt.Fprintf(c.App.Writer, "- %s from %s to %s\n", humanSize, dir, outDir)
	}

	humanTotalSize := humanize.IBytes(uint64(totalSize))
	fmt.Fprintf(c.App.Writer, "Total of %s.  Confirm (y/N)? ", humanTotalSize)
	var confirm string
	fmt.Fscanf(c.App.Reader, "%s", &confirm)
	if confirm != "y" {
		return nil
	}

	// Only record the first error from here on out.
	updateError := func(in error) (ok bool) {
		ok = in == nil
		if err != nil {
			return
		}
		err = in
		return
	}

	needSave := false
	for i, dir := range dirs {
		project := projects[i]
		outSuperDir := outSuperDirs[i]
		outDir := outDirs[i]

		fmt.Fprintf(c.App.Writer, "\n[%d/%d] Copying from %s...\n", i+1, len(dirs), dir)

		// Use rsync to copy the files.
		args := []string{
			"-ah",
			"--progress",
			"-r",
			dir,
			outSuperDir,
		}
		cmd := exec.Command("/usr/bin/rsync", args...)
		cmd.Stdin = c.App.Reader
		cmd.Stdout = c.App.Writer
		cmd.Stderr = c.App.ErrWriter
		if !updateError(cmd.Run()) {
			continue
		}

		// Now rename the .extras dir (if it exists)
		extrasPath := filepath.Join(outDir, ".extras")
		var hasExtrasDir bool
		if _, statErr := os.Stat(extrasPath); statErr == nil {
			hasExtrasDir = true
		} else if !errors.Is(statErr, fs.ErrNotExist) {
			updateError(statErr)
			continue
		}
		if hasExtrasDir {
			newExtrasPath := filepath.Join(outDir, "extras")
			fmt.Fprintf(c.App.Writer, "Renaming extras dir... ")
			if !updateError(os.Rename(extrasPath, newExtrasPath)) {
				continue
			}
			fmt.Fprintf(c.App.Writer, "done.\n")
		} else {
			fmt.Fprintln(c.App.Writer, "No extras dir.")
		}

		// finally, update the stage in the project struct (if we got here)
		project.Stage = psPushed
		needSave = true
	}

	if needSave {
		updateError(save())
	}
	updateError(resume())

	return err
}

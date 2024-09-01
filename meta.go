package main

// spell-checker:ignore urfave

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	cli "github.com/urfave/cli/v2"
)

func cmdCfgNew() *cli.Command {
	return &cli.Command{
		Name:        "new",
		Usage:       "create a new project",
		Description: "Creates a new movie or tv show project.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "type",
				Usage:    "The type of the ripping project, either 'movie' or 'show'.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "The name of the ripping project.",
				Required: true,
			},
		},
		Action: cmdNew,
	}
}

func cmdNew(c *cli.Context) error {
	tp, ts, save, err := ripCmdInit(c)
	if err != nil {
		return err
	}
	name := c.String("name")
	if _, found := ts.FindByName(name); found {
		return fmt.Errorf("there is already a project named %s", name)
	}
	var t projectType
	switch c.String("type") {
	case "movie":
		t = ptMovie
	case "show":
		t = ptShow
	default:
		return fmt.Errorf("invalid value of --type")
	}
	ps := &projectState{
		Name:  name,
		Pt:    t,
		Stage: psWorking,
	}
	ts.Projects = append(ts.Projects, ps)

	projectDir, err := tp.TmmProjectDir(ps)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("could not create project dir %s: %w", projectDir, err)
	}

	return save()
}

func cmdCfgFinish() *cli.Command {
	return &cli.Command{
		Name:      "finish",
		Usage:     "finish any projects that have been pushed",
		ArgsUsage: " ", // Makes help text a bit nicer.
		Action:    cmdFinish,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "y",
				Usage: "always confirms interactive prompts",
				Value: false,
			},
		},
	}
}

func cmdFinish(c *cli.Context) error {
	tp, ts, save, err := ripCmdInit(c)
	if err != nil {
		return err
	}
	found := ts.FindByStage(psPushed)
	if len(found) == 0 {
		return errors.New("no projects are marked as pushed")
	}
	dirs := make([]string, 0, len(found))
	for _, p := range found {
		projectDir, err := tp.TmmProjectDir(p)
		if err != nil {
			return fmt.Errorf("could not get project dir for %s", p.Name)
		}
		dirs = append(dirs, projectDir)
	}
	if !c.Bool("y") {
		fmt.Fprintf(c.App.Writer, "Will delete the following directories:\n")
		for _, d := range dirs {
			fmt.Fprintf(c.App.Writer, "- %s\n", d)
		}
		fmt.Fprintf(c.App.Writer, "Confirm (y/N)? ")
		var confirm string
		fmt.Fscanf(c.App.Reader, "%s", &confirm)
		if confirm != "y" {
			return nil
		}
	}

	updateError := func(in error) (inOk bool) {
		inOk = in == nil
		if err != nil {
			return
		}
		err = in
		return
	}
	type empty struct{}
	removed := make(map[*projectState]empty)
	for i, p := range found {
		d := dirs[i]
		if updateError(os.RemoveAll(d)) {
			removed[p] = empty{}
		}
	}
	newProjects := make([]*projectState, 0, len(ts.Projects))
	for _, oldProject := range ts.Projects {
		if _, found := removed[oldProject]; found {
			continue
		}
		newProjects = append(newProjects, oldProject)
	}
	ts.Projects = newProjects
	return save()
}

func cmdCfgStage() *cli.Command {
	return &cli.Command{
		Name:  "stage",
		Usage: "show or change the stage of a ripping project.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "The name of the ripping project.",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "working",
				Usage: "move the project to the working stage.",
			},
			&cli.BoolFlag{
				Name:  "ready",
				Usage: "move the project to the ready-for-push stage.",
			},
			&cli.BoolFlag{
				Name:  "pushed",
				Usage: "move the project to the pushed stage.",
			},
		},
		Action: cmdStage,
	}
}

func cmdStage(c *cli.Context) error {
	_, ts, save, err := ripCmdInit(c)
	if err != nil {
		return err
	}

	name := c.String("name")
	project, found := ts.FindByName(name)
	if !found {
		return fmt.Errorf("no project named %s", name)
	}

	working := c.Bool("working")
	ready := c.Bool("ready")
	pushed := c.Bool("pushed")

	switch {
	case working && !ready && !pushed:
		project.Stage = psWorking
	case !working && ready && !pushed:
		project.Stage = psReadyForPush
	case !working && !ready && pushed:
		project.Stage = psPushed
	case !working && !ready && !pushed:
		// Nothing to do, but not an error
	default:
		return errors.New("invalid combination of --working --ready --pushed")
	}
	if err := save(); err != nil {
		return err
	}
	fmt.Fprintf(c.App.Writer, "%s\n", project.Stage)
	return nil
}

func cmdCfgRipLs() *cli.Command {
	return &cli.Command{
		Name:   "ls",
		Usage:  "show all ripping projects.",
		Action: cmdRipLs,
	}
}

func cmdRipLs(c *cli.Context) error {
	_, ts, _, err := ripCmdInit(c)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(c.App.Writer, 0, 4, 3, byte(' '), 0)
	fmt.Fprintln(tw, "name\ttype\tstage\tdir override")
	fmt.Fprintln(tw, "----\t----\t-----\t------------")
	for _, p := range ts.Projects {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", p.Name, p.Pt, p.Stage, p.TmmDirOverride)
	}
	return tw.Flush()
}

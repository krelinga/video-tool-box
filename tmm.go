package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
)

func cmdCfgTmm() *cli.Command {
	return &cli.Command{
		Name:  "tmm",
		Usage: "run Tiny Media Manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "The name of the project to open TMM for.",
				Required: true,
			},
		},
		Action: cmdTmm,
	}
}

func runTmmAndWait() error {
	cmd := exec.Command("open", "-W", "/Applications/tinyMediaManager.app/")
	return cmd.Run()
}

func findFlagFile(base, fileName string) (string, error) {
	s := make([]string, 1)
	s[0] = base
	for len(s) > 0 {
		lastIdx := len(s) - 1
		c := s[lastIdx]
		s = s[:lastIdx]

		entries, err := os.ReadDir(c)
		if err != nil {
			return "", fmt.Errorf("Could not read dir %s: %w", c, err)
		}
		for _, e := range entries {
			cPath := filepath.Join(c, e.Name())
			if e.Name() == fileName {
				return cPath, nil
			} else if e.IsDir() {
				s = append(s, cPath)
			}
			// the file is uninteresting, do nothing.
		}
	}
	return "", fmt.Errorf("Could not find a file named %s under %s", fileName, base)
}

func cmdTmm(c *cli.Context) error {
	tp, ok := toolPathsFromContext(c.Context)
	if !ok {
		return errors.New("toolPaths not present in context")
	}
	ts, err := readToolState(tp.StatePath())
	if err != nil {
		return err
	}

	name := c.String("name")
	project, found := ts.FindByName(name)
	if !found {
		return fmt.Errorf("No project named %s", name)
	}

	if project.Stage != psWorking {
		return fmt.Errorf("Project %s is not in the working stage: %s", name, project.Stage)
	}

	flagFile := fmt.Sprintf(".%d", rand.Int31())
	projectDir, err := tp.TmmProjectDir(project)
	if err != nil {
		return err
	}
	flagPath := filepath.Join(projectDir, flagFile)
	if err := os.WriteFile(flagPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("Could not create flag path %s: %w", flagPath, err)
	}

	if err := runTmmAndWait(); err != nil {
		return fmt.Errorf("Could not run TMM: %w", err)
	}

	var base string
	switch project.Pt {
	case ptUndef:
		return errors.New("Undefined project state")
	case ptMovie:
		base = tp.TmmMoviesDir()
	case ptShow:
		base = tp.TmmShowsDir()
	default:
		return fmt.Errorf("Unexpected project state: %v", project.Pt)
	}
	newFlagPath, err := findFlagFile(base, flagFile)
	if err != nil {
		return err
	}

	project.TmmDirOverride = filepath.Dir(newFlagPath)
	if err := writeToolState(ts, tp.StatePath()); err != nil {
		return err
	}

	if err := os.Remove(newFlagPath); err != nil {
		return fmt.Errorf("Could not remove new flag path %s: %w", newFlagPath, err)
	}

	return nil
}

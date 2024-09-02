package main

// spell-checker:ignore urfave .tcprofile tvshow.nfo

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

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
			return "", fmt.Errorf("could not read dir %s: %w", c, err)
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
	return "", fmt.Errorf("could not find a file named %s under %s", fileName, base)
}

type nfoFileInfo struct {
	path    string
	content string
}

func findNfoFiles(base string) ([]*nfoFileInfo, error) {
	var files []*nfoFileInfo

	// Recursively walk the base directory and find all .nfo files.
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".nfo" && filepath.Base(path) != "tvshow.nfo" {
			files = append(files, &nfoFileInfo{path: path})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Read the contents of all .nfo files in parallel.
	var wg sync.WaitGroup
	errorChan := make(chan error, len(files))
	for _, file := range files {
		file := file
		wg.Add(1)
		go func() {
			defer wg.Done()

			content, err := os.ReadFile(file.path)
			if err != nil {
				errorChan <- fmt.Errorf("error reading file %s: %v", file.path, err)
				return
			}

			file.content = string(content)
		}()
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	for err := range errorChan {
		return nil, err
	}

	return files, nil
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
		return fmt.Errorf("no project named %s", name)
	}

	if project.Stage != psWorking {
		return fmt.Errorf("project %s is not in the working stage: %s", name, project.Stage)
	}

	flagFile := fmt.Sprintf(".%d", rand.Int31())
	projectDir, err := tp.TmmProjectDir(project)
	if err != nil {
		return err
	}
	flagPath := filepath.Join(projectDir, flagFile)
	if err := os.WriteFile(flagPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("could not create flag path %s: %w", flagPath, err)
	}

	// find the .nfo files that exist right now and a hash of their contents.
	previousNfoFiles, err := findNfoFiles(projectDir)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.App.Writer, "found %d .nfo already-existing files\n", len(previousNfoFiles))

	if err := runTmmAndWait(); err != nil {
		return fmt.Errorf("could not run TMM: %w", err)
	}

	var base string
	switch project.Pt {
	case ptUndef:
		return errors.New("undefined project state")
	case ptMovie:
		base = tp.TmmMoviesDir()
	case ptShow:
		base = tp.TmmShowsDir()
	default:
		return fmt.Errorf("unexpected project state: %v", project.Pt)
	}
	newFlagPath, err := findFlagFile(base, flagFile)
	if err != nil {
		return err
	}

	projectDir = filepath.Dir(newFlagPath)
	project.TmmDirOverride = projectDir
	if err := writeToolState(ts, tp.StatePath()); err != nil {
		return err
	}
	if err := os.Remove(newFlagPath); err != nil {
		return fmt.Errorf("could not remove new flag path %s: %w", newFlagPath, err)
	}

	// Find the .nfo files and their contents that exist now, and update any
	// .tcprofile files that need to change.
	newNfoFiles, err := findNfoFiles(projectDir)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.App.Writer, "found %d .nfo new files\n", len(newNfoFiles))

	return nil
}

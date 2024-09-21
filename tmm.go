package main

// spell-checker:ignore urfave .tcprofile tvshow.nfo Tcprofile

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/krelinga/go-lib/chans"
	"github.com/krelinga/video-tool-box/nfo"
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

func nfoPathToTcprofilePath(nfoPath string) string {
	return strings.TrimSuffix(nfoPath, ".nfo") + ".tcprofile"
}

func crawlNfoFiles(base string) (chan string, chan error) {
	nfoFiles := make(chan string)
	errors := make(chan error)
	go func() {
		defer close(nfoFiles)
		defer close(errors)
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				errors <- err
				return nil
			}
			if !d.IsDir() && filepath.Ext(path) == ".nfo" && filepath.Base(path) != "tvshow.nfo" {
				nfoFiles <- path
			}
			return nil
		})
		if err != nil {
			errors <- err
		}
	}()
	return nfoFiles, errors
}

func readNfoFile(path string) (*nfoFileInfo, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", path, err)
	}
	return &nfoFileInfo{path: path, content: string(content)}, nil
}

func each[inType any, outType any](f func(inType) outType, in ...inType) []outType {
	out := make([]outType, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func goWait(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}

func goWaitAll(fs ...func()) {
	wg := sync.WaitGroup{}
	for _, f := range fs {
		goWait(&wg, f)
	}
	wg.Wait()
}

func findNfoFiles(base string) ([]*nfoFileInfo, error) {
	nfoPaths, pathErrors := crawlNfoFiles(base)
	nfoFiles, readErrors := chans.ParallelErr(20, nfoPaths, readNfoFile)

	finalInfos := []*nfoFileInfo{}
	var finalErr error
	goWaitAll(
		func() {
			for info := range nfoFiles {
				finalInfos = append(finalInfos, info)
			}
		},
		func() {
			for err := range chans.Merge(pathErrors, readErrors) {
				if finalErr == nil {
					finalErr = err
				}
			}
		},
	)

	return finalInfos, finalErr
}

func guessTranscodeProfile(in *nfo.Content) string {
	if in.Width == 1920 && in.Height == 1080 {
		return "hd"
	}
	findSubstring := func(target string, in []string) bool {
		target = strings.ToLower(target)
		for _, s := range in {
			s = strings.ToLower(s)
			if strings.Contains(s, target) {
				return true
			}
		}
		return false
	}
	notations := in.Genres
	notations = append(notations, in.Tags...)
	if findSubstring("anime", notations) || findSubstring("animation", notations) {
		return "sd_animation"
	}
	return "sd_live_action"
}

func updateTranscodeProfiles(old, new []*nfoFileInfo) error {
	makeMap := func(files []*nfoFileInfo) map[string]*nfoFileInfo {
		m := make(map[string]*nfoFileInfo, len(files))
		for _, f := range files {
			m[f.path] = f
		}
		return m
	}
	oldMap := makeMap(old)
	updateNeeded := []*nfoFileInfo{}
	for _, newFile := range new {
		oldFile, found := oldMap[newFile.path]
		if !found || oldFile.content != newFile.content {
			updateNeeded = append(updateNeeded, newFile)
		}
	}

	for _, file := range updateNeeded {
		nfoFile := file.path
		showNfo, err := nfo.Parse(nfoFile, strings.NewReader(file.content))
		if err != nil {
			return fmt.Errorf("could not parse NFO file %s: %w", nfoFile, err)
		}
		profile := guessTranscodeProfile(showNfo)
		profileFile := nfoPathToTcprofilePath(nfoFile)
		if err := os.WriteFile(profileFile, []byte(profile+"\n"), 0644); err != nil {
			return fmt.Errorf("could not write profile file %s: %w", profileFile, err)
		}
	}

	return nil
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

	if err := updateTranscodeProfiles(previousNfoFiles, newNfoFiles); err != nil {
		return err
	}

	return nil
}

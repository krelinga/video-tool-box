package main

// spell-checker:ignore urfave Tcprofiles .tcprofile writetcprofiles Tcprofile

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/krelinga/video-tool-box/nfo"
	cli "github.com/urfave/cli/v2"
)

func cmdCfgWriteTcprofiles() *cli.Command {
	return &cli.Command{
		Name:  "writetcprofiles",
		Usage: "Write any missing .tcprofile files under the given directory.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "dir",
				Usage:    "The directory to search for .tcprofile files.",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Do not write any files, just print what would be done.",
			},
		},
		Action: cmdWritetcprofiles,
	}
}

func cmdWritetcprofiles(c *cli.Context) error {
	files, err := findNfoFiles(c.String("dir"))
	if err != nil {
		return err
	}

	type counters struct {
		alreadyExists int
		added         int
		total         int
	}

	progressChan := make(chan counters)
	progressDoneChan := make(chan struct{})
	go func() {
		const updateInterval = 10
		defer close(progressDoneChan)
		totals := counters{}
		output := func() {
			fmt.Fprintf(c.App.Writer, "Processed %d/%d files.  %d already exist, %d added.\n", totals.total, len(files), totals.alreadyExists, totals.added)
		}
		for p := range progressChan {
			totals.alreadyExists += p.alreadyExists
			totals.added += p.added
			totals.total += p.total
			if totals.total%updateInterval == 0 {
				output()
			}
		}
		output()
	}()

	const parallelism = 20
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup
	errorChan := make(chan error, len(files))
	for _, f := range files {
		f := f
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			one_count := counters{
				total: 1,
			}
			defer wg.Done()
			defer func() { <-sem }()
			defer func() { progressChan <- one_count }()

			tcProfilePath := nfoPathToTcprofilePath(f.path)
			if _, err := os.Stat(tcProfilePath); err == nil {
				// file already exists, so we can skip it.
				one_count.alreadyExists += 1
				return
			}

			nfoInfo, err := nfo.Parse(f.path, strings.NewReader(f.content))
			if err != nil {
				errorChan <- fmt.Errorf("could not parse nfo file %s: %w", f.path, err)
				return
			}
			if !c.Bool("dry-run") {
				guessedTcProfile := guessTranscodeProfile(nfoInfo)
				profileFile := nfoPathToTcprofilePath(f.path)
				if err := os.WriteFile(profileFile, []byte(guessedTcProfile+"\n"), 0644); err != nil {
					errorChan <- fmt.Errorf("could not write tcprofile file %s: %w", profileFile, err)
					return
				}
			}
			one_count.added += 1
		}()
	}

	wg.Wait()
	close(progressChan)
	<-progressDoneChan
	close(errorChan)

	for err := range errorChan {
		return err
	}

	return nil
}

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
		},
		Action: cmdWritetcprofiles,
	}
}

func cmdWritetcprofiles(c *cli.Context) error {
	files, err := findNfoFiles(c.String("dir"))
	if err != nil {
		return err
	}

	progressChan := make(chan struct{})
	progressDoneChan := make(chan struct{})
	go func() {
		const updateInterval = 10
		updates := 0
		defer close(progressDoneChan)
		for range progressChan {
			updates += 1
			if updates%updateInterval == 0 {
				fmt.Fprintf(c.App.Writer, "Processed %d/%d files\n", updates, len(files))
			}
		}
	}()

	const parallelism = 20
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup
	errorChan := make(chan error, len(files))
	for _, f := range files {
		f := f
		wg.Add(1)
		go func() {
			sem <- struct{}{}
			defer wg.Done()
			defer func() { <-sem }()
			defer func() { progressChan <- struct{}{} }()

			tcProfilePath := nfoPathToTcprofilePath(f.path)
			if _, err := os.Stat(tcProfilePath); err == nil {
				// file already exists, so we can skip it.
				return
			}

			nfoInfo, err := nfo.Parse(f.path, strings.NewReader(f.content))
			if err != nil {
				errorChan <- fmt.Errorf("could not parse nfo file %s: %w", f.path, err)
				return
			}
			guessedTcProfile := guessTranscodeProfile(nfoInfo)
			profileFile := nfoPathToTcprofilePath(f.path)
			if err := os.WriteFile(profileFile, []byte(guessedTcProfile+"\n"), 0644); err != nil {
				errorChan <- fmt.Errorf("could not write tcprofile file %s: %w", profileFile, err)
				return
			}
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

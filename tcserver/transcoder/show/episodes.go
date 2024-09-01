package show

import (
	"os"
	"path/filepath"
	"regexp"
)

var seasonRE = regexp.MustCompile(`Season \d+`)

func FindEpisodes(dir string) ([]string, error) {
	// Find the season directories.
	filesInDir, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	seasons := []string{}
	for _, file := range filesInDir {
		if seasonRE.MatchString(file.Name()) {
			seasons = append(seasons, file.Name())
		}
	}

	episodes := []string{}
	// Find the episodes in each season directory
	for _, season := range seasons {
		seasonFiles, err := os.ReadDir(filepath.Join(dir, season))
		if err != nil {
			return nil, err
		}
		for _, file := range seasonFiles {
			if filepath.Ext(file.Name()) == ".mkv" {
				episodes = append(episodes, filepath.Join(dir, season, file.Name()))
			}
		}
	}
	return episodes, nil
}

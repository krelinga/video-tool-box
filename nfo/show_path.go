package nfo

// spell-checker:ignore tvshow.nfo

import (
	"errors"
	"path/filepath"
)

func showNfoPath(episodeNfoPath string) (string, error) {
	if detectFileContext(episodeNfoPath) != Episode {
		return "", errors.New("file is not an episode NFO")
	}

	seasonDir, _ := filepath.Split(filepath.Clean(episodeNfoPath))
	if seasonDir == "" {
		return "", errors.New("invalid path")
	}
	showDir, _ := filepath.Split(filepath.Clean(seasonDir))
	return filepath.Join(showDir, "tvshow.nfo"), nil
}

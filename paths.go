package main

import (
	"context"
	"errors"
	"path/filepath"
)

type toolPaths struct {
	homeDir     string
	currentDir  string
	nasMountDir string
}

func newProdToolPaths() (*toolPaths, error) {
	tp := &toolPaths{}
	var err error
	tp.currentDir, err = getEnvVar("PWD")
	if err != nil {
		return tp, err
	}
	tp.homeDir, err = getEnvVar("HOME")
	if err != nil {
		return tp, err
	}
	tp.nasMountDir, err = getEnvVar("VTB_NAS_MOUNT_DIR")
	if err != nil {
		return tp, err
	}
	return tp, nil
}

func (tp *toolPaths) HomeDir() string {
	return tp.homeDir
}

func (tp *toolPaths) CurrentDir() string {
	return tp.currentDir
}

func (tp *toolPaths) NasMountDir() string {
	return tp.nasMountDir
}

func (tp *toolPaths) MoviesDir() string {
	return filepath.Join(tp.HomeDir(), "Movies")
}

func (tp *toolPaths) TmmMoviesDir() string {
	return filepath.Join(tp.MoviesDir(), "tmm_movies")
}

func (tp *toolPaths) TmmShowsDir() string {
	return filepath.Join(tp.MoviesDir(), "tmm_shows")
}

func (tp *toolPaths) StatePath() string {
	return filepath.Join(tp.HomeDir(), ".vtb_state")
}

func (tp *toolPaths) TmmProjectDir(ps *projectState) (string, error) {
	if len(ps.Name) == 0 {
		return "", errors.New("empty Name field in toolState")
	}
	if len(ps.TmmDirOverride) > 0 {
		return ps.TmmDirOverride, nil
	}
	switch ps.Pt {
	case ptMovie:
		return filepath.Join(tp.TmmMoviesDir(), ps.Name), nil
	case ptShow:
		return filepath.Join(tp.TmmShowsDir(), ps.Name), nil
	}
	return "", errors.New("unexpected value of ps.Pt")
}

func (tp *toolPaths) TmmProjectExtrasDir(ps *projectState) (string, error) {
	projectDir, err := tp.TmmProjectDir(ps)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectDir, ".extras"), nil
}

func (tp *toolPaths) ConfigPath() string {
	return filepath.Join(tp.HomeDir(), ".vtb_config.json")
}

var toolPathsContextKey string = "toolPathsContextKey"

func newToolPathsContext(ctx context.Context, tp *toolPaths) context.Context {
	return context.WithValue(ctx, toolPathsContextKey, tp)
}

func toolPathsFromContext(ctx context.Context) (*toolPaths, bool) {
	value, ok := ctx.Value(toolPathsContextKey).(*toolPaths)
	return value, ok
}

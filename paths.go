package main

import (
    "context"
    "errors"
    "path/filepath"
)


type toolPaths struct {
    homeDir     string
    currentDir  string
    nasMountDir      string
}

func newProdToolPaths() (toolPaths, error) {
    tp := toolPaths{}
    var err error
    tp.currentDir, err = getEnvVar("PWD")
    if err != nil { return tp, err }
    tp.homeDir, err = getEnvVar("HOME")
    if err != nil { return tp, err }
    tp.nasMountDir, err = getEnvVar("VTB_NAS_MOUNT_DIR")
    if err != nil { return tp, err }
    return tp, nil
}

func (tp toolPaths) HomeDir() string {
    return tp.homeDir
}

func (tp toolPaths) CurrentDir() string {
    return tp.currentDir
}

func (tp toolPaths) NasMountDir() string {
    return tp.nasMountDir
}

func (tp toolPaths) MoviesDir() string {
    return filepath.Join(tp.HomeDir(), "Movies")
}

func (tp toolPaths) TmmMoviesDir() string {
    return filepath.Join(tp.MoviesDir(), "tmm_movies")
}

func (tp toolPaths) TmmShowsDir() string {
    return filepath.Join(tp.MoviesDir(), "tmm_shows")
}

func (tp toolPaths) StatePath() string {
    return filepath.Join(tp.HomeDir(), ".vtb_state")
}

func (tp toolPaths) TmmProjectDir(ts toolState) (string, error) {
    if len(ts.Name) == 0 {
        return "", errors.New("Empty Name field in toolState")
    }
    switch ts.Pt {
    case ptMovie:
        return filepath.Join(tp.TmmMoviesDir(), ts.Name), nil
    case ptShow:
        return filepath.Join(tp.TmmShowsDir(), ts.Name), nil
    }
    return "", errors.New("Unexpected value of ts.Pt")
}

func (tp toolPaths) TmmProjectExtrasDir(ts toolState) (string, error) {
    projectDir, err := tp.TmmProjectDir(ts)
    if err != nil { return "", err }
    return filepath.Join(projectDir, ".extras"), nil
}

var toolPathsContextKey string = "toolPathsContextKey"

func newToolPathsContext(ctx context.Context, tp toolPaths) context.Context {
    return context.WithValue(ctx, toolPathsContextKey, tp)
}

func toolPathsFromContext(ctx context.Context) (toolPaths, bool) {
    value, ok := ctx.Value(toolPathsContextKey).(toolPaths)
    return value, ok
}

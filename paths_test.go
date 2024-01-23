package main

import (
    "os"
    "testing"
)

func setEnvVar(t *testing.T, key string, value string) {
    t.Helper()
    if err := os.Setenv(key, value); err != nil {
        t.Fatalf("could not set env var %s to %s, error was %s", key, value, err)
    }
}

func TestEmptyHOMEAndPWD(t *testing.T) {
    setEnvVar(t, "HOME", "")
    setEnvVar(t, "PWD", "")
    _, err := newProdToolPaths()
    if err == nil {
        t.Error("Expected error from newProdToolPaths()")
    }
}

func TestEmptyHOME(t *testing.T) {
    setEnvVar(t, "HOME", "")
    setEnvVar(t, "PWD", "/workdir")
    _, err := newProdToolPaths()
    if err == nil {
        t.Error("Expected error from newProdToolPaths()")
    }
}

func TestEmptyPWD(t *testing.T) {
    setEnvVar(t, "HOME", "/homedir")
    setEnvVar(t, "PWD", "")
    _, err := newProdToolPaths()
    if err == nil {
        t.Error("Expected error from newProdToolPaths()")
    }
}

func TestBasicPaths(t *testing.T) {
    setEnvVar(t, "HOME", "/homedir")
    setEnvVar(t, "PWD", "/workdir")
    toolPaths, err := newProdToolPaths()
    if err != nil {
        t.Fatalf("could not create toolPaths: %s", err)
    }
    if toolPaths.HomeDir() != "/homedir" {
        t.Error(toolPaths.HomeDir())
    }
    if toolPaths.CurrentDir() != "/workdir" {
        t.Error(toolPaths.CurrentDir())
    }
    if toolPaths.MoviesDir() != "/homedir/Movies" {
        t.Error(toolPaths.CurrentDir())
    }
    if toolPaths.TmmMoviesDir() != "/homedir/Movies/tmm_movies" {
        t.Error(toolPaths.TmmMoviesDir())
    }
    if toolPaths.TmmShowsDir() != "/homedir/Movies/tmm_shows" {
        t.Error(toolPaths.TmmShowsDir())
    }
    if toolPaths.StatePath() != "/homedir/.vtb_state" {
        t.Error(toolPaths.StatePath())
    }
}

func TestTmmProjectDir(t *testing.T) {
    setEnvVar(t, "HOME", "/homedir")
    setEnvVar(t, "PWD", "/workdir")
    toolPaths, err := newProdToolPaths()
    if err != nil {
        t.Fatalf("could not create toolPaths: %s", err)
    }

    type testCase struct {
        Name string
        Ts toolState
        Err bool
        Path string
    }
    testCases := []testCase{
        {
            Name: "empty name",
            Ts: toolState{
                Pt: ptMovie,
            },
            Err: true,
        },
        {
            Name: "undefined project type",
            Ts: toolState{
                Name: "some name",
                Pt: ptUndef,
            },
            Err: true,
        },
        {
            Name: "movie project type",
            Ts: toolState{
                Name: "some name",
                Pt: ptMovie,
            },
            Path: "/homedir/Movies/tmm_movies/some name",
        },
        {
            Name: "tv show project type",
            Ts: toolState{
                Name: "some name",
                Pt: ptShow,
            },
            Path: "/homedir/Movies/tmm_shows/some name",
        },
    }

    for _, tc := range testCases {
        path, err := toolPaths.TmmProjectDir(tc.Ts)
        if tc.Err && err == nil {
            t.Errorf("%s: expected error but got none", tc.Name)
        }
        if !tc.Err {
            if err != nil {
                t.Errorf("%s: unexpected error %s", tc.Name, err)
            }
            if len(tc.Path) > 0 && path != tc.Path {
                t.Errorf("%s: expected path %s, got %s", tc.Name, tc.Path, path)
            }
        }
    }
}

func TestTmmProjectExtrasDir(t *testing.T) {
    setEnvVar(t, "HOME", "/homedir")
    setEnvVar(t, "PWD", "/workdir")
    toolPaths, err := newProdToolPaths()
    if err != nil {
        t.Fatalf("could not create toolPaths: %s", err)
    }

    type testCase struct {
        Name string
        Ts toolState
        Err bool
        Path string
    }
    testCases := []testCase{
        {
            Name: "empty name",
            Ts: toolState{
                Pt: ptMovie,
            },
            Err: true,
        },
        {
            Name: "undefined project type",
            Ts: toolState{
                Name: "some name",
                Pt: ptUndef,
            },
            Err: true,
        },
        {
            Name: "movie project type",
            Ts: toolState{
                Name: "some name",
                Pt: ptMovie,
            },
            Path: "/homedir/Movies/tmm_movies/some name/.extras",
        },
        {
            Name: "tv show project type",
            Ts: toolState{
                Name: "some name",
                Pt: ptShow,
            },
            Path: "/homedir/Movies/tmm_shows/some name/.extras",
        },
    }

    for _, tc := range testCases {
        path, err := toolPaths.TmmProjectExtrasDir(tc.Ts)
        if tc.Err && err == nil {
            t.Errorf("%s: expected error but got none", tc.Name)
        }
        if !tc.Err {
            if err != nil {
                t.Errorf("%s: unexpected error %s", tc.Name, err)
            }
            if len(tc.Path) > 0 && path != tc.Path {
                t.Errorf("%s: expected path %s, got %s", tc.Name, tc.Path, path)
            }
        }
    }
}

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

func TestNewProdToolPaths(t *testing.T) {
    // This manipulates environment variables, so it needs to be run serially.
    oldHome := os.Getenv("HOME")
    oldPwd := os.Getenv("PWD")
    defer func() {
        setEnvVar(t, "HOME", oldHome)
        setEnvVar(t, "PWD", oldPwd)
    }()

    type testCase struct {
        Name    string
        Home    string
        Pwd     string
        ExpectError bool
    }
    cases := []testCase{
        {
            Name: "No Homedir or PWD",
            ExpectError: true,
        },
        {
            Name: "No Homedir",
            Pwd: "/workingdir",
            ExpectError: true,
        },
        {
            Name: "No PWD",
            Home: "/homedir",
            ExpectError: true,
        },
        {
            Name: "Homedir and PWD set",
            Home: "/homedir",
            Pwd: "/workingdir",
        },
    }
    for _, tc := range cases {
        t.Run(tc.Name, func(t *testing.T) {
            setEnvVar(t, "HOME", tc.Home)
            setEnvVar(t, "PWD", tc.Pwd)
            _, err := newProdToolPaths()
            if tc.ExpectError && err == nil {
                t.Error("Expected error but didn't get one.")
            }
            if !tc.ExpectError && err != nil {
                t.Errorf("didn't expect an error, but got one: %v", err)
            }
        })
    }
}

func TestBasicPaths(t *testing.T) {
    setEnvVar(t, "HOME", "/homedir")
    setEnvVar(t, "PWD", "/workdir")
    tp, err := newProdToolPaths()
    if err != nil {
        t.Fatalf("could not create tp: %s", err)
    }
    if tp.HomeDir() != "/homedir" {
        t.Error(tp.HomeDir())
    }
    if tp.CurrentDir() != "/workdir" {
        t.Error(tp.CurrentDir())
    }
    if tp.MoviesDir() != "/homedir/Movies" {
        t.Error(tp.CurrentDir())
    }
    if tp.TmmMoviesDir() != "/homedir/Movies/tmm_movies" {
        t.Error(tp.TmmMoviesDir())
    }
    if tp.TmmShowsDir() != "/homedir/Movies/tmm_shows" {
        t.Error(tp.TmmShowsDir())
    }
    if tp.StatePath() != "/homedir/.vtb_state" {
        t.Error(tp.StatePath())
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

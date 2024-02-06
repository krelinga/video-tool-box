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
    oldVtbNasMountDir := os.Getenv("VTB_NAS_MOUNT_DIR")
    defer func() {
        setEnvVar(t, "HOME", oldHome)
        setEnvVar(t, "PWD", oldPwd)
        setEnvVar(t, "VTB_NAS_MOUNT_DIR", oldVtbNasMountDir)
    }()

    type testCase struct {
        Name    string
        Home    string
        Pwd     string
        NasMount     string
        ExpectError bool
    }
    cases := []testCase{
        {
            Name: "No Homedir, PWD, or NAS",
            ExpectError: true,
        },
        {
            Name: "No Homedir",
            Pwd: "/workingdir",
            NasMount: "/nas",
            ExpectError: true,
        },
        {
            Name: "No PWD",
            Home: "/homedir",
            NasMount: "/nas",
            ExpectError: true,
        },
        {
            Name: "No NAS",
            Home: "/homedir",
            Pwd: "/workingdir",
            ExpectError: true,
        },
        {
            Name: "Homedir, PWD, and NAS set",
            Home: "/homedir",
            Pwd: "/workingdir",
            NasMount: "/nas",
        },
    }
    for _, tc := range cases {
        t.Run(tc.Name, func(t *testing.T) {
            setEnvVar(t, "HOME", tc.Home)
            setEnvVar(t, "PWD", tc.Pwd)
            setEnvVar(t, "VTB_NAS_MOUNT_DIR", tc.NasMount)
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
    t.Parallel()
    tp := toolPaths{
        homeDir: "/homedir",
        currentDir: "/workdir",
        nasMountDir: "/nas",
    }
    if tp.HomeDir() != "/homedir" {
        t.Error(tp.HomeDir())
    }
    if tp.CurrentDir() != "/workdir" {
        t.Error(tp.CurrentDir())
    }
    if tp.NasMountDir() != "/nas" {
        t.Error(tp.NasMountDir())
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
    t.Parallel()
    tp := toolPaths{
        homeDir: "/homedir",
        currentDir: "/workdir",
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
        t.Run(tc.Name, func(t *testing.T) {
            t.Parallel()
            path, err := tp.TmmProjectDir(tc.Ts)
            if tc.Err && err == nil {
                t.Error("expected error but got none")
            }
            if !tc.Err {
                if err != nil {
                    t.Errorf("unexpected error %s", err)
                }
                if len(tc.Path) > 0 && path != tc.Path {
                    t.Errorf("expected path %s, got %s", tc.Path, path)
                }
            }
        })
    }
}

func TestTmmProjectExtrasDir(t *testing.T) {
    t.Parallel()
    tp := toolPaths{
        homeDir: "/homedir",
        currentDir: "/workdir",
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
        t.Run(tc.Name, func(t *testing.T) {
            t.Parallel()
            path, err := tp.TmmProjectExtrasDir(tc.Ts)
            if tc.Err && err == nil {
                t.Error("expected error but got none")
            }
            if !tc.Err {
                if err != nil {
                    t.Errorf("unexpected error %s", err)
                }
                if len(tc.Path) > 0 && path != tc.Path {
                    t.Errorf("expected path %s, got %s", tc.Path, path)
                }
            }
        })
    }
}

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
    oldVtbNasCanonDir := os.Getenv("VTB_NAS_CANON_DIR")
    defer func() {
        setEnvVar(t, "HOME", oldHome)
        setEnvVar(t, "PWD", oldPwd)
        setEnvVar(t, "VTB_NAS_MOUNT_DIR", oldVtbNasMountDir)
        setEnvVar(t, "VTB_NAS_CANON_DIR", oldVtbNasCanonDir)
    }()

    type testCase struct {
        Name        string
        Home        string
        Pwd         string
        NasMount    string
        NasCanon    string
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
            NasCanon: "smb://nas",
            ExpectError: true,
        },
        {
            Name: "No PWD",
            Home: "/homedir",
            NasMount: "/nas",
            NasCanon: "smb://nas",
            ExpectError: true,
        },
        {
            Name: "No NAS Mount",
            Home: "/homedir",
            Pwd: "/workingdir",
            NasCanon: "smb://nas",
            ExpectError: true,
        },
        {
            Name: "No NAS Canon",
            Home: "/homedir",
            Pwd: "/workingdir",
            NasMount: "/nas",
            ExpectError: true,
        },
        {
            Name: "Homedir, PWD, NAS Mount and NAS Canon set",
            Home: "/homedir",
            Pwd: "/workingdir",
            NasMount: "/nas",
            NasCanon: "smb://nas",
        },
    }
    for _, tc := range cases {
        t.Run(tc.Name, func(t *testing.T) {
            setEnvVar(t, "HOME", tc.Home)
            setEnvVar(t, "PWD", tc.Pwd)
            setEnvVar(t, "VTB_NAS_MOUNT_DIR", tc.NasMount)
            setEnvVar(t, "VTB_NAS_CANON_DIR", tc.NasCanon)
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
    tp := &toolPaths{
        homeDir: "/homedir",
        currentDir: "/workdir",
        nasMountDir: "/nas",
        nasCanonDir: "smb://nas",
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
    if tp.NasCanonDir() != "smb://nas" {
        t.Error(tp.NasCanonDir())
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
    if tp.ConfigPath() != "/homedir/.vtb_config.json" {
        t.Error(tp.ConfigPath())
    }
}

func TestTmmProjectDir(t *testing.T) {
    t.Parallel()
    tp := &toolPaths{
        homeDir: "/homedir",
        currentDir: "/workdir",
    }

    type testCase struct {
        Name string
        Ts *toolState
        Err bool
        Path string
    }
    testCases := []testCase{
        {
            Name: "empty name",
            Ts: &toolState{
                Pt: ptMovie,
            },
            Err: true,
        },
        {
            Name: "undefined project type",
            Ts: &toolState{
                Name: "some name",
                Pt: ptUndef,
            },
            Err: true,
        },
        {
            Name: "movie project type",
            Ts: &toolState{
                Name: "some name",
                Pt: ptMovie,
            },
            Path: "/homedir/Movies/tmm_movies/some name",
        },
        {
            Name: "tv show project type",
            Ts: &toolState{
                Name: "some name",
                Pt: ptShow,
            },
            Path: "/homedir/Movies/tmm_shows/some name",
        },
        {
            Name: "Override Set",
            Ts: &toolState{
                Name: "some name",
                TmmDirOverride: "/foo/bar",
            },
            Path: "/foo/bar",
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
    tp := &toolPaths{
        homeDir: "/homedir",
        currentDir: "/workdir",
    }

    type testCase struct {
        Name string
        Ts *toolState
        Err bool
        Path string
    }
    testCases := []testCase{
        {
            Name: "empty name",
            Ts: &toolState{
                Pt: ptMovie,
            },
            Err: true,
        },
        {
            Name: "undefined project type",
            Ts: &toolState{
                Name: "some name",
                Pt: ptUndef,
            },
            Err: true,
        },
        {
            Name: "movie project type",
            Ts: &toolState{
                Name: "some name",
                Pt: ptMovie,
            },
            Path: "/homedir/Movies/tmm_movies/some name/.extras",
        },
        {
            Name: "tv show project type",
            Ts: &toolState{
                Name: "some name",
                Pt: ptShow,
            },
            Path: "/homedir/Movies/tmm_shows/some name/.extras",
        },
        {
            Name: "Override Set",
            Ts: &toolState{
                Name: "some name",
                TmmDirOverride: "/foo/bar",
            },
            Path: "/foo/bar/.extras",
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

func TestTranslateNasDir(t *testing.T) {
    tp := &toolPaths {
        nasMountDir: "/nas",
        nasCanonDir: "smb://nas",
    }

    type testCase struct {
        Name   string
        Path   string
        Expect string
        Err    bool
    }
    testCases := []testCase {
        {
            Name: "Prefix Matched",
            Path: "/nas/some/file",
            Expect: "smb://nas/some/file",
        },
        {
            Name: "Prefix Not Matched",
            Path: "/x/some/file",
            Err: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.Name, func (t *testing.T) {
            t.Parallel()
            newPath, err := tp.TranslateNasDir(tc.Path)
            if len(tc.Expect) > 0 && tc.Expect != newPath {
                t.Error(newPath)
            }
            if tc.Err && err == nil {
                t.Error("Expected error but didn't get it.")
            }
            if !tc.Err && err != nil {
                t.Errorf("Unexpected error %s", err)
            }
        })
    }

}

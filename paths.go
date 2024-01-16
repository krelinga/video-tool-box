package main

import (
    "log"
    "os"
    "path/filepath"
)

var (
    homeDir = func() string {
        home := os.Getenv("HOME")
        if len(home) == 0 {
            log.Fatal("could not read HOME env var")
        }
        return home
    }()
    currentDir = func() string {
        pwd := os.Getenv("PWD")
        if len(pwd) == 0 {
            log.Fatal("could not read PWD env var")
        }
        return pwd
    }()
    moviesDir = filepath.Join(homeDir, "Movies")
    tmmMoviesDir = filepath.Join(moviesDir, "tmm_movies")
    tmmShowsDir = filepath.Join(moviesDir, "tmm_shows")
    statePath = filepath.Join(homeDir, ".vtb_state")
)

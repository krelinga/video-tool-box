package main

import (
    "bytes"
    "context"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

type testApp struct {
    paths toolPaths

    stdin bytes.Buffer
    stdout bytes.Buffer
    stderr bytes.Buffer
}

func newTestApp(t *testing.T) *testApp {
    t.Helper()
    return &testApp{
        paths: toolPaths{
            homeDir: setUpTempDir(t),
            currentDir: setUpTempDir(t),
        },
    }
}

func (ta *testApp) Delete(t *testing.T) {
    t.Helper()
    tearDownTempDir(t, ta.Paths().HomeDir())
    tearDownTempDir(t, ta.Paths().CurrentDir())
}

func (ta *testApp) Paths() toolPaths {
    return ta.paths
}

func (ta *testApp) Stdin() *bytes.Buffer {
    return &ta.stdin
}

func (ta *testApp) Stdout() *bytes.Buffer {
    return &ta.stdout
}

func (ta *testApp) Stderr() *bytes.Buffer {
    return &ta.stderr
}

func (ta *testApp) Reset() {
    for _, b := range []*bytes.Buffer{&ta.stdin, &ta.stdout, &ta.stderr} {
        b.Reset()
    }
}

func (ta *testApp) Run(args... string) error {
    app := appCfg()
    app.Reader = &ta.stdin
    app.Writer = &ta.stdout
    app.ErrWriter = &ta.stderr
    ctx := newToolPathsContext(context.Background(), ta.paths)
    fullArgs := []string{"vtb"}
    fullArgs = append(fullArgs, args...)
    return app.RunContext(ctx, fullArgs)
}

func TestRootHelp(t *testing.T) {
    t.Parallel()
    ta := newTestApp(t)
    defer ta.Delete(t)

    if err := ta.Run("help"); err != nil {
        t.Errorf("unexpected error: %s", err)
    }
    if ta.Stderr().Len() > 0 {
        t.Errorf("unexpected stderr output: %s", ta.Stderr().String())
    }
    stdout := ta.Stdout().String()
    hasName := strings.Contains(stdout, "NAME:")
    hasUsage := strings.Contains(stdout, "\nUSAGE:")
    hasCommands := strings.Contains(stdout, "\nCOMMANDS:")
    if !hasName || !hasUsage || !hasCommands {
        t.Errorf("Output does not look like help text:\n%s", stdout)
    }
}

func TestRipSequence(t *testing.T) {
    t.Parallel()
    ta := newTestApp(t)
    defer ta.Delete(t)

    // Set up a fake DVD directory.
    mkdir := func(path string) {
        if err := os.Mkdir(path, 0755); err != nil {
            t.Fatalf("error creating directory %s: %s", path, err)
        }
    }
    mkdir(ta.Paths().MoviesDir())
    mkdir(ta.Paths().TmmMoviesDir())
    writeTestFile := func(name string) {
        path := filepath.Join(ta.Paths().CurrentDir(), name)
        if err := os.WriteFile(path, []byte(name), 0644); err != nil {
            t.Fatalf("error creating fake MKV file %s: %s", path, err)
        }
    }
    // Prefix is to control the sorted order more-easily.
    writeTestFile("a_title.mkv")
    writeTestFile("b_extra.mkv")
    writeTestFile("c_skip.mkv")
    writeTestFile("d_delete.mkv")

    runNoError := func(args... string) bool {
        if err := ta.Run(args...); err != nil {
            t.Errorf("Error running with args %v: %s", args, err)
            return false
        }
        return true
    }
    testMetaUnconfigured := func() {
        if !runNoError("meta") {
            return
        }
        initMetaOut := ta.Stdout().String()
        noProjectConfigured := "no project configured"
        if !strings.Contains(initMetaOut, noProjectConfigured) {
            t.Errorf("Unexpected 'meta' output for no configured project: %s", initMetaOut)
        }
        ta.Reset()
    }
    testMetaConfigured := func() {
        if !runNoError("meta") {
            return
        }
        afterNewMetaOut := ta.Stdout().String()
        if !strings.Contains(afterNewMetaOut, "Test Movie") {
            t.Errorf("Unexpected 'meta' output for configured project: %s", afterNewMetaOut)
        }
        ta.Reset()
    }

    testMetaUnconfigured()

    // Project names don't have to be quoted on the shell, so we pass
    // "Test" and "Movie" as two separate strings here.
    if !runNoError("new", "movie", "Test", "Movie") {
        return
    }
    ta.Reset()

    testMetaConfigured()

    if _, err := ta.Stdin().WriteString("t\nx\ns\nd\n"); err != nil {
        t.Fatalf("error writing to test stdin: %s", err)
    }
    if !runNoError("dir") {
        return
    }
    if leftover := ta.Stdin().Len(); leftover > 0 {
        t.Errorf("dir has %d leftover bytes in stdin", leftover)
    }
    check := func(path string) {
        expContents := filepath.Base(path)
        actBytes, err := os.ReadFile(path)
        if err != nil {
            t.Errorf("Could not read %s: %s", path, err)
            return
        }
        actContents := string(actBytes)
        if expContents != actContents {
            t.Errorf("Bad contents in %s: expected \"%s\", actual \"%s\"", path, expContents, actContents)
        }
    }
    tmmMovieDir := filepath.Join(ta.Paths().TmmMoviesDir(), "Test Movie")
    check(filepath.Join(tmmMovieDir, "a_title.mkv"))
    check(filepath.Join(tmmMovieDir, ".extras", "b_extra.mkv"))
    check(filepath.Join(ta.Paths().CurrentDir(), "c_skip.mkv"))
    files, err := os.ReadDir(ta.Paths().CurrentDir())
    if err != nil {
        t.Errorf("could not read current directory: %s", err)
        return
    }
    if len(files) != 1 {
        t.Errorf("Expected 1 entry in working directory, found %d", len(files))
    }
    ta.Reset()

    testMetaConfigured()

    if !runNoError("finish") {
        return
    }
    ta.Reset()

    testMetaUnconfigured()
}

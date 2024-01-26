package main

import (
    "bytes"
    "context"
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

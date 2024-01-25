package main

import (
    "bytes"
    "context"
    "strings"
    "testing"
)

type testApp struct {
    workDir string
    homeDir string

    stdin bytes.Buffer
    stdout bytes.Buffer
    stderr bytes.Buffer
}

func newTestApp(t *testing.T) *testApp {
    return &testApp{
        workDir: setUpTempDir(t),
        homeDir: setUpTempDir(t),
    }
}

func (ta *testApp) Delete(t *testing.T) {
    tearDownTempDir(t, ta.workDir)
    tearDownTempDir(t, ta.homeDir)
}

func (ta *testApp) WorkDir() string {
    return ta.workDir
}

func (ta *testApp) HomeDir() string {
    return ta.homeDir
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
    tp := toolPaths{
        homeDir: ta.homeDir,
        currentDir: ta.workDir,
    }
    ctx := newToolPathsContext(context.Background(), tp)
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

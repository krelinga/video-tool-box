package main

import (
    "context"
    "log"
    "os"
)

func main() {
    internal := func() error {
        tp, err := newProdToolPaths()
        if err != nil {
            return err
        }
        ctx := newToolPathsContext(context.Background(), tp)
        app := appCfg()
        app.Reader = os.Stdin
        app.Writer = os.Stdout
        app.ErrWriter = os.Stderr
        return app.RunContext(ctx, os.Args)
    }
    if err := internal(); err != nil {
        log.Fatal(err)
    }
}

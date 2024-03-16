package main

import (
    "errors"
    "fmt"
    "path/filepath"
    "strings"

    cli "github.com/urfave/cli/v2"
    "github.com/krelinga/video-tool-box/pb"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func cmdCfgInfo() *cli.Command {
    return &cli.Command{
        Name: "info",
        Usage: "talk to MKV Info Server",
        Action: cmdInfo,
    }
}

func cmdInfo(c *cli.Context) error {
    tp, ok := toolPathsFromContext(c.Context)
    if !ok {
        return errors.New("toolPaths not present in context")
    }

    args := c.Args().Slice()
    if len(args) != 1 {
        return errors.New("Expected a single file path")
    }
    path := args[0]

    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("Could not get absolute path of %s: %w", path, err)
    }

    pathSuffix, found := strings.CutPrefix(absPath, tp.MoviesDir())
    if !found {
        return fmt.Errorf("Path %s is not in movies dir %s", absPath, tp.MoviesDir())
    }
    pathSuffix = strings.TrimLeft(pathSuffix, "/")

    const serverPathPrefeix = "/Movies"
    serverPath := filepath.Join(serverPathPrefeix, pathSuffix)

    fmt.Fprintf(c.App.Writer, "Will call server for %s\n", absPath)
    fmt.Fprintf(c.App.Writer, "serverPath: %s\n", serverPath)

    serverTarget, err := getEnvVar("VTB_MKVINFOSERVER_TARGET")
    if err != nil {
        return err
    }

    conn, err := grpc.DialContext(c.Context, serverTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return fmt.Errorf("Error dialing %s: %w", serverTarget, err)
    }
    defer conn.Close()

    client := pb.NewMkvInfoClient(conn)

    request := &pb.GetMkvInfoRequest{
        In: serverPath,
    }
    reply, err := client.GetMkvInfo(c.Context, request)
    if err != nil {
        return fmt.Errorf("Error sending request %v to server: %w", request, err)
    }

    fmt.Fprintf(c.App.Writer, "reply: %v\n", reply)

    return nil
}

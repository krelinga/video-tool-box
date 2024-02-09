package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "strconv"

    "github.com/krelinga/video-tool-box/pb"
    "google.golang.org/grpc"
)

func getPort() (int, error) {
    const envVar = "VTB_TCSERVER_PORT"
    portString, err := getEnvVar(envVar)
    if err != nil {
        return 0, err
    }
    port, err := strconv.Atoi(portString)
    if err != nil {
        return 0, fmt.Errorf("env var %s should be a port number, saw %s", envVar, portString)
    }
    return port, nil
}

func listVideoPaths(path string) error {
    newPath, err := translatePath(path)
    if err != nil {
        return err
    }

    entries, err := os.ReadDir(newPath)
    if err != nil {
        return err
    }
    for _, entry := range entries {
        fmt.Println(entry.Name())
    }
    return nil
}

func mainOrError() error {
    fmt.Println("hello world!")
    if err := listVideoPaths("smb://truenas/media"); err != nil {
        return err
    }
    demoHbParser := func() error {
        canonPath := "smb://truenas/media/Experimental/5sec.mkv.stdout"
        realPath, err := translatePath(canonPath)
        if err != nil {
            return err
        }
        stdoutFile, err := os.Open(realPath)
        if err != nil {
            return err
        }
        for entry := range parseHbOutput(stdoutFile) {
            fmt.Printf("Progress: {%s}\n", entry)
        }
        return nil
    }
    if err := demoHbParser(); err != nil {
        return err
    }
    port, err := getPort()
    if err != nil {
        return err
    }
    statePath, err := getEnvVar("VTB_TCSERVER_STATE_PATH")
    if err != nil {
        return err
    }
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return err
    }
    grpcServer := grpc.NewServer()
    pb.RegisterTCServerServer(grpcServer, newTcServer(statePath))
    grpcServer.Serve(lis)  // Runs as long as the server is alive.

    return nil
}

func main() {
    if err := mainOrError(); err != nil {
        log.Fatal(err)
    }
}

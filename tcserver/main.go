package main

import (
    "fmt"
    "log"
    "net"
    "strconv"

    "github.com/krelinga/video-tool-box/pb"
    "github.com/krelinga/video-tool-box/tcserver/transcoder"
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

func mainOrError() error {
    fmt.Println("hello world!")
    port, err := getPort()
    if err != nil {
        return err
    }
    profile, err := getEnvVar("VTB_TCSERVER_PROFILE")
    if err != nil {
        return err
    }
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return err
    }
    grpcServer := grpc.NewServer()
    tran := transcoder.Transcoder{
        FileWorkers: 1,
        MaxQueuedFiles: 10000,
        ShowWorkers: 1,
        MaxQueuedShows: 1,
    }
    if err := tran.Start(); err != nil {
        return err
    }
    defer tran.Stop()
    pb.RegisterTCServerServer(grpcServer, newTcServer(profile, &tran))
    grpcServer.Serve(lis)  // Runs as long as the server is alive.

    return nil
}

func main() {
    if err := mainOrError(); err != nil {
        log.Fatal(err)
    }
}

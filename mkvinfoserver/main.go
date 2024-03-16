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

func getEnvVar(name string) (string, error) {
    value := os.Getenv(name)
    if len(value) == 0 {
        return "", fmt.Errorf("env var %s is not set", name)
    }
    return value, nil
}

func getPort() (int, error) {
    const envVar = "VTB_MKVINFOSERVER_PORT"
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
    fmt.Println("hello world from mkvinfoserver!")
    port, err := getPort()
    if err != nil {
        return err
    }
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return err
    }
    grpcServer := grpc.NewServer()
    pb.RegisterMkvInfoServer(grpcServer, &MkvInfoServer{})
    grpcServer.Serve(lis)  // Runs as long as the server is alive.

    return nil
}

func main() {
    if err := mainOrError(); err != nil {
        log.Fatal(err)
    }
}

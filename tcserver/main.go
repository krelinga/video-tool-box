package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "net"
    "os"
    "strconv"
    "strings"

    "github.com/krelinga/video-tool-box/pb"
    "google.golang.org/grpc"
)

type tcServer struct {
    pb.UnimplementedTCServerServer
}

func (tcs *tcServer) HelloWorld(ctx context.Context, req *pb.HelloWorldRequest) (*pb.HelloWorldReply, error) {
    fmt.Printf("Saw request: %v\n", req)
    rep := &pb.HelloWorldReply{
        Out: req.In,
    }
    return rep, nil
}

func getEnvVar(name string) (string, error) {
    value := os.Getenv(name)
    if len(value) == 0 {
        return "", errors.New(fmt.Sprintf("env var %s is not set", name))
    }
    return value, nil
}

func getPort() (int, error) {
    const envVar = "VTB_TCSERVER_PORT"
    portString, err := getEnvVar(envVar)
    if err != nil {
        return 0, err
    }
    port, err := strconv.Atoi(portString)
    if err != nil {
        return 0, errors.New(fmt.Sprintf("env var %s should be a port number, saw %s", envVar, portString))
    }
    return port, nil
}

func listVideoPaths(path string) error {
    oldPrefix, err := getEnvVar("VTB_TCSERVER_IN_PATH_PREFIX")
    if err != nil {
        return err
    }
    newPrefix, err := getEnvVar("VTB_TCSERVER_OUT_PATH_PREFIX")
    if err != nil {
        return err
    }

    cutPath, found := strings.CutPrefix(path, oldPrefix)
    if !found {
        return errors.New(fmt.Sprintf("path %s does not start with prefix %s", path, oldPrefix))
    }
    newPath := newPrefix + cutPath

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
    port, err := getPort()
    if err != nil {
        return err
    }
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return err
    }
    grpcServer := grpc.NewServer()
    pb.RegisterTCServerServer(grpcServer, &tcServer{})
    grpcServer.Serve(lis)  // Runs as long as the server is alive.

    return nil
}

func main() {
    if err := mainOrError(); err != nil {
        log.Fatal(err)
    }
}

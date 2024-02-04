package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "net"
    "os"
    "strconv"

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

func getPort() (int, error) {
    const envVar = "VTB_TCSERVER_PORT"
    portString := os.Getenv(envVar)
    if len(portString) == 0 {
        return 0, errors.New(fmt.Sprintf("set %s env var", envVar))
    }
    port, err := strconv.Atoi(portString)
    if err != nil {
        return 0, errors.New(fmt.Sprintf("env var %s should be a port number, saw %s", envVar, portString))
    }
    return port, nil
}
func mainOrError() error {
    fmt.Println("hello world!")
    port, err := getPort()
    if err != nil {
        return err
    }
    lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
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

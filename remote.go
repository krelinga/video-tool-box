package main

import (
    "fmt"

    cli "github.com/urfave/cli/v2"
    "github.com/krelinga/video-tool-box/pb"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func cmdCfgRemote() *cli.Command {
    return &cli.Command{
        Name: "remote",
        Usage: "talk to remote transcoding server.",
        Action: cmdRemote,
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "target",
                Usage: "rpc target to talk to.",
                Required: true,
            },
        },
    }
}

func cmdRemote(c *cli.Context) error {
    conn, err := grpc.DialContext(c.Context, c.String("target"), grpc.WithTransportCredentials(insecure.NewCredentials()))
    if  err != nil {
        return err
    }
    defer conn.Close()

    client := pb.NewTCServerClient(conn)

    resp, err := client.HelloWorld(c.Context, &pb.HelloWorldRequest{In: "taters"})
    if err != nil {
        return err
    }
    fmt.Println(resp)
    return nil
}

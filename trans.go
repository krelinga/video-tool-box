package main

import (
    "errors"
    "fmt"
    "os"
    "os/exec"

    cli "github.com/urfave/cli/v2"
)

type handbrakeFlags []string

var gHandbrakeProfile = map[string]handbrakeFlags{
    "mkv_h265_1080p30": {
        "-Z", "Matroska/H.265 MKV 1080p30",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
}

func cmdCfgTrans() *cli.Command {
    return &cli.Command{
        Name: "trans",
        Usage: "transcode video",
        Action: cmdTrans,
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name: "input",
                Usage: "input path to read & transcode.",
                Required: true,
            },
            &cli.StringFlag{
                Name: "output",
                Usage: "output path to write transcoded file to.",
                Required: true,
            },
            &cli.StringFlag{
                Name: "profile",
                Usage: "name of the transcoding profile to use.",
                Required: true,
            },
        },
    }
}

func cmdTrans(c *cli.Context) error {
    handbrake := c.String("handbrake")
    if len(handbrake) == 0 {
        return errors.New("'trans' command only available when --handbrake is set")
    }

    input := c.String("input")
    output := c.String("output")
    profile := c.String("profile")

    profileFlags, ok := gHandbrakeProfile[profile]
    if !ok {
        return errors.New(fmt.Sprintf("unknown profile %s", profile))
    }
    standardFlags := []string{
        "-i", input,
        "-o", output,
    }
    cmd := exec.Command(handbrake)
    cmd.Args = append(cmd.Args, standardFlags...)
    cmd.Args = append(cmd.Args, profileFlags...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    fmt.Println("starting Handbrake....")
    if err := cmd.Run(); err != nil { return err }
    fmt.Println("...handbrake finished")

    return nil
}

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

func cmdTrans() *cli.Command {
    fn := func(c *cli.Context) error {
        handbrake := c.String("handbrake")
        if len(handbrake) == 0 {
            return errors.New("'trans' command only available when --handbrake is set")
        }

        input, err := getEnvVar("VTB_INPUT")
        if err != nil { return err }
        output, err := getEnvVar("VTB_OUTPUT")
        if err != nil { return err }
        profile, err := getEnvVar("VTB_PROFILE")
        if err != nil { return err }

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

    return &cli.Command{
        Name: "trans",
        Usage: "transcode video",
        Action: fn,
    }
}

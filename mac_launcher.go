package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
)

func main() {
    fmt.Println("Hello from launcher!")

    cmd := exec.Command("docker")

    ext := func(a *[]string, args ...string) {
        for _, arg := range args {
            *a = append(*a, arg)
        }
    }
    ext(&cmd.Args, "run")
    ext(&cmd.Args, "--pull", "always")
    ext(&cmd.Args, "krelinga/video-tool-box")

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    if err := cmd.Run(); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Docker run successful.")
}

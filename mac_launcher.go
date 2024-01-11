package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
)

func main() {
    fmt.Println("Hello from launcher!")

    docker_cmd := exec.Command("docker", "run", "--pull", "always", "krelinga/video-tool-box")
    docker_cmd.Stdout = os.Stdout
    docker_cmd.Stderr = os.Stderr
    docker_cmd.Stdin = os.Stdin
    if err := docker_cmd.Run(); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Docker run successful.")
}

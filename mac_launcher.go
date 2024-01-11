package main

import (
    "fmt"
    "log"
    "os/exec"
)

func main() {
    fmt.Println("Hello from launcher!")

    docker_cmd := exec.Command("docker", "run", "--pull", "always", "krelinga/video-tool-box")
    if err := docker_cmd.Run(); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Docker run successful.")
}

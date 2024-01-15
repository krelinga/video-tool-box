package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
)

func main() {
    fmt.Println("Hello from launcher!")

    cmd := exec.Command("docker")

    ext := func(args ...string) {
        for _, arg := range args {
            cmd.Args = append(cmd.Args, arg)
        }
    }
    ext("run")
    ext("--pull", "always")
    home := os.Getenv("HOME")
    movies := filepath.Join(home, "Movies")
    const workdir string = "/workdir"
    if len(home) == 0 {
        log.Print("could not read HOME")
    } else {
        ext("-v", fmt.Sprintf("%s:%s", movies, workdir))
    }
    ext("krelinga/video-tool-box")
    if len(home) != 0 {
        ext("--work_dir", workdir)
    }

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    log.Print("final docker command line:", cmd.Args)
    if err := cmd.Run(); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Docker run successful.")
}

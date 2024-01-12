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
    pwd := os.Getenv("PWD")
    const workdir string = "/workdir"
    if len(pwd) == 0 {
        log.Print("could not read PWD")
    } else {
        ext(&cmd.Args, "-v", fmt.Sprintf("%s:%s", pwd, workdir))
    }
    ext(&cmd.Args, "krelinga/video-tool-box")
    if len(pwd) != 0 {
        ext(&cmd.Args, "--work_dir", workdir)
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

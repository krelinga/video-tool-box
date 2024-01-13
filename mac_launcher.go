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

    ext := func(args ...string) {
        for _, arg := range args {
            cmd.Args = append(cmd.Args, arg)
        }
    }
    ext("run")
    ext("--pull", "always")
    pwd := os.Getenv("PWD")
    const workdir string = "/workdir"
    if len(pwd) == 0 {
        log.Print("could not read PWD")
    } else {
        ext("-v", fmt.Sprintf("%s:%s", pwd, workdir))
    }
    tmm_movies_host_dir := os.Getenv("VTB_TMM_MOVIES_DIR")
    const tmm_movies string = "/tmm_movies"
    if len(tmm_movies_host_dir) == 0 {
        log.Print("could not read VTB_TMM_MOVIES_DIR")
    } else {
        ext("-v", fmt.Sprintf("%s:%s", tmm_movies_host_dir, tmm_movies))
    }
    ext("krelinga/video-tool-box")
    if len(pwd) != 0 {
        ext("--work_dir", workdir)
    }
    if len(tmm_movies_host_dir) != 0 {
        ext("--tmm_movies", tmm_movies)
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

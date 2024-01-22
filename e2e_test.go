package main

import (
    "fmt"
    "testing"
    "os/exec"

    "github.com/google/uuid"
)

func buildContainer(t *testing.T, containerId string) {
    t.Helper()
    cmd := exec.Command("docker", "image", "build", "-t", containerId, ".")
    if err := cmd.Run(); err != nil {
        t.Fatalf("could not build docker container: %s", err)
    }
    t.Log("Finished building docker container.")
}

func deleteContainer(t *testing.T, containerId string) {
    t.Helper()
    cmd := exec.Command("docker", "image", "rm", containerId)
    if err := cmd.Run(); err != nil {
        t.Fatalf("could not delete docker container: %s", err)
    }
    t.Log("Finished deleting docker container.")
}

func runVtb(containerId string, args... string) error {
    cmd := exec.Command("docker", "run", "--rm", "-t", containerId)
    cmd.Args = append(cmd.Args, args...)
    return cmd.Run()
}

func TestDockerBuildAndRun(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping test in short mode.")
    }

    containerId := fmt.Sprintf("vtb-e2e-test-%s", uuid.NewString())
    buildContainer(t, containerId)
    defer deleteContainer(t, containerId)

    if err := runVtb(containerId); err != nil {
        t.Errorf("error running vtb: %s", err)
    }
}

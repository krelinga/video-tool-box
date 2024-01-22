package main

import (
    "fmt"
    "testing"
    "os/exec"

    "github.com/google/uuid"
)

type testContainer struct {
    containerId string
}

func newTestContainer() testContainer {
    return testContainer{
        containerId: fmt.Sprintf("vtb-e2e-test-%s", uuid.NewString()),
    }
}


func (tc testContainer) Build(t *testing.T) {
    t.Helper()
    cmd := exec.Command("docker", "image", "build", "-t", tc.containerId, ".")
    if err := cmd.Run(); err != nil {
        t.Fatalf("could not build docker container: %s", err)
    }
    t.Log("Finished building docker container.")
}

func (tc testContainer) Delete(t *testing.T) {
    t.Helper()
    cmd := exec.Command("docker", "image", "rm", tc.containerId)
    if err := cmd.Run(); err != nil {
        t.Fatalf("could not delete docker container: %s", err)
    }
    t.Log("Finished deleting docker container.")
}

func (tc testContainer) Run(args... string) error {
    cmd := exec.Command("docker", "run", "--rm", "-t", tc.containerId)
    cmd.Args = append(cmd.Args, args...)
    return cmd.Run()
}

func testDockerBuildAndRun(t *testing.T, tc testContainer) {
    if err := tc.Run(); err != nil {
        t.Errorf("error running vtb: %s", err)
    }
}

func TestE2E(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping end-to-end test in short mode.")
    }
    tc := newTestContainer()
    tc.Build(t)

    t.Run("Docker Build & Run", func(t *testing.T) {
        testDockerBuildAndRun(t, tc)
    })

    defer tc.Delete(t)
}

// +build e2e

package main

import (
    "bytes"
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

func captureOutput(cmd *exec.Cmd) *bytes.Buffer {
    cmdOutput := &bytes.Buffer{}
    cmd.Stdout = cmdOutput
    cmd.Stderr = cmdOutput
    return cmdOutput
}

func (tc testContainer) Build(t *testing.T) {
    t.Helper()
    cmd := exec.Command("docker", "image", "build", "-t", tc.containerId, ".")
    cmdOutput := captureOutput(cmd)
    if err := cmd.Run(); err != nil {
        t.Fatalf("could not build docker container: %s %s", err, cmdOutput)
    }
    t.Log("Finished building docker container.")
}

func (tc testContainer) Delete(t *testing.T) {
    t.Helper()
    cmd := exec.Command("docker", "image", "rm", tc.containerId)
    cmdOutput := captureOutput(cmd)
    if err := cmd.Run(); err != nil {
        t.Fatalf("could not delete docker container: %s %s", err, cmdOutput)
    }
    t.Log("Finished deleting docker container.")
}

func (tc testContainer) Run(args... string) (*bytes.Buffer, error) {
    cmd := exec.Command("docker", "run", "--rm", "-t", tc.containerId)
    cmdOutput := captureOutput(cmd)
    cmd.Args = append(cmd.Args, args...)
    return cmdOutput, cmd.Run()
}

func testDockerBuildAndRun(t *testing.T, tc testContainer) {
    if output, err := tc.Run(); err != nil {
        t.Errorf("error running vtb: %s %s", err, output)
    }
}

func TestE2E(t *testing.T) {
    t.Parallel()
    tc := newTestContainer()
    tc.Build(t)
    defer tc.Delete(t)

    t.Run("Docker Build & Run", func(t *testing.T) {
        testDockerBuildAndRun(t, tc)
    })

}

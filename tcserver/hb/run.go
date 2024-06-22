package hb

import (
    "errors"
    "fmt"
    "io"
    "os"
    "os/exec"
)

func Run(inPath, outPath, profile string, prog func(*Progress)) error {
    flags, err := GetFlags(profile, inPath, outPath)
    if err != nil {
        return err
    }

    if _, err := os.Stat(outPath); !errors.Is(err, os.ErrNotExist) {
        return errors.New("Output file already exists")
    }

    stdOutPath := outPath + ".stdout"
    stdOutFile, err := os.Create(stdOutPath)
    if err != nil {
        return fmt.Errorf("could not open %s: %v", stdOutPath, err)
    }
    defer stdOutFile.Close()

    stdErrPath := outPath + ".stderr"
    stdErrFile, err := os.Create(stdErrPath)
    if err != nil {
        return fmt.Errorf("could not open %s: %v", stdErrPath, err)
    }
    defer stdErrFile.Close()

    // A pipe to allow stdout to be consumed via a Reader.
    hbPipeReader, hbPipeWriter := io.Pipe()

    // Tee the output of Handbrake so that it goes to both stdOutFile and
    // progressReader
    progressReader := io.TeeReader(hbPipeReader, stdOutFile)

    // parse entries out of progressReader and into a channel.
    progressCh := ParseOutput(progressReader)

    // Consume from progressCh while Handbrake is running, and update s.
    // Notify progressDone when all updates have been consumed.
    progressDone := make(chan struct{})
    go func() {
        for u := range(progressCh) {
            prog(u)
        }
        progressDone <- struct{}{}
    }()

    cmd := exec.Command("HandBrakeCLI")
    cmd.Args = flags
    cmd.Stdin = os.Stdin
    cmd.Stdout = hbPipeWriter
    cmd.Stderr = stdErrFile
    err = cmd.Run()

    // Wait for all progress to be processed before we exit.  This also makes
    // sure that all output from Handbrake was written to stdOutFile via the
    // above tee.  Note that the tee does not close stdOutFile when hbPipeWriter
    // is closed, so we rely on a `defer stdOutFile.Close()` statement above.
    hbPipeWriter.Close()
    <- progressDone

    return err
}

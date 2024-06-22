package main

import (
    "errors"
    "sync"

    "github.com/krelinga/video-tool-box/tcserver/hb"
)

type newState int
const (
    newStateUnknown newState = iota
    newStateNotStarted
    newStateInProgress
    newStateComplete
    newStateError
)

type newSingleFileState struct {
    // Read-only after the struct is created.
    inPath string
    outPath string
    profile string
    onDone func()  // mu will not be held when this is called.

    // Read/written concurrently.
    latest *hb.Progress
    state newState
    err error

    // Must be held when reading or writing latest, state, or err.
    mu *sync.Mutex
}

// Transcodes sfs.inPath into sfs.outPath according to sfs.profile.
//
// Blocks until transcoding is finished, returning any error.
func (sfs *newSingleFileState) Transcode() error {
    // TODO: implement.
    return nil
}

func newTranscodeFileWorker(in <-chan *newSingleFileState) {
    for work := range in {
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            work.state = newStateInProgress
        }()
        err := work.Transcode()
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            if err != nil {
                work.err = err
                work.state = newStateError
            } else {
                work.state = newStateComplete
            }
        }()
        if work.onDone != nil {
            work.onDone()
        }
    }
}

type newShowState struct {
    // Read-only after the struct is created.
    inDirPath string
    outParentDirPath string
    profile string

    // Read/written concurrently.
    fileStates []*newSingleFileState
    state newState
    err error

    mu sync.Mutex
}

// Discovers all the .mkv files under ss.inDirPath, and transcodes them.
//
// A corresponding output directory is created under ss.outParentDirPath to
// store the output.  The transcoding makes as much progress as possible, and
// does not stop as soon as any error is encountered.
//
// Any returned error only captures errors that happened during setup.  Any
// error that happens during the transcoding of individual files is stored in
// ss.fileStates.
//
// The work of transcoding individual files is delegated to fileQueue
func (ss *newShowState) Transcode(fileQueue chan<- *newSingleFileState) error {
    // TODO: check if the output path already exists.
    // TODO: discover .mkv files & map in file paths to out file paths.
    mkvMap := make(map[string]string)
    // TODO: copy over any show-level files.
    wg := sync.WaitGroup{}
    wg.Add(len(mkvMap))
    for fromPath, toPath := range mkvMap {
        sfs := &newSingleFileState{
            inPath: fromPath,
            outPath: toPath,
            profile: ss.profile,
            onDone: func() {
                wg.Done()
            },
            mu: &ss.mu,
        }
        fileQueue <- sfs
    }
    wg.Wait()

    return nil
}

func newTranscodeShowWorker(in <-chan *newShowState, fileQueue chan<- *newSingleFileState) {
    for work := range in {
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            work.state = newStateInProgress
        }()
        err := work.Transcode(fileQueue)
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            if err != nil {
                work.err = err
                work.state = newStateError
            } else {
                work.state = newStateComplete
            }
        }()
    }
}

var (
    newTranscoderStoppedErr = errors.New("Transcoder has been stopped")
    newTranscoderAlreadyExistsErr = errors.New("Transcode already exists for this name")
    newTranscoderAlreadyStartedErr = errors.New("Transcoder already started.")
    newTranscoderInvalidConfigErr = errors.New("Transcoder config is invalid.")
    newTranscoderNotStartedErr = errors.New("Transcoder not started.")
    newTranscoderFullErr = errors.New("Transcoder is full.")
    newTranscoderNotExistErr = errors.New("Transcode does not exist.")
)

type newTranscoder struct {
    // Exposed configuration variables.
    FileWorkers int
    MaxQueuedFiles int
    ShowWorkers int
    MaxQueuedShows int

    // State management.
    started bool
    // Close when it's time to stop all processing.
    stop chan struct{}

    // processing queues.
    fileQueue chan *newSingleFileState
    showQueue chan *newShowState

    files map[string]*newSingleFileState
    shows map[string]*newShowState
    mu sync.Mutex
}

func (t *newTranscoder) Start() error {
    t.mu.Lock()
    defer t.mu.Unlock()
    select {
    case <- t.stop:
        return newTranscoderStoppedErr
    }
    if t.started {
        return newTranscoderAlreadyStartedErr
    }
    filesCfgValid := t.FileWorkers >= 1 && t.MaxQueuedFiles >= 1
    showsCfgValid := t.ShowWorkers >= 1 && t.MaxQueuedShows >= 1
    if !(filesCfgValid && showsCfgValid) {
        return newTranscoderInvalidConfigErr
    }
    t.fileQueue = make(chan *newSingleFileState, t.MaxQueuedFiles)
    for i := 0; i < t.FileWorkers; i++ {
        go newTranscodeFileWorker(t.fileQueue)
    }
    t.showQueue = make(chan *newShowState, t.MaxQueuedShows)
    for i := 0; i < t.ShowWorkers; i++ {
        go newTranscodeShowWorker(t.showQueue, t.fileQueue)
    }
    go func() {
        <- t.stop
        close(t.fileQueue)
        close(t.showQueue)
    }()
    return nil
}

func (t *newTranscoder) Stop() {
    t.mu.Lock()
    defer t.mu.Unlock()

    if !t.started {
        panic("Transcoder not started")
    }
    select {
    case <- t.stop:
        return
    default:
        close(t.stop)
    }
}

func (t *newTranscoder) StartFile(name, inPath, outPath, profile string) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    if !t.started {
        return newTranscoderNotStartedErr
    }
    state, found := t.files[name]
    if found && (state.state == newStateInProgress || state.state == newStateNotStarted) {
        return newTranscoderAlreadyExistsErr
    }
    state = &newSingleFileState{
        inPath: inPath,
        outPath: outPath,
        profile: profile,
        state: newStateNotStarted,
        mu: &sync.Mutex{},
    }
    select {
    case <- t.stop:
        return newTranscoderStoppedErr
    case t.fileQueue <- state:
        t.files[name] = state
        return nil
    default:
        return newTranscoderFullErr
    }
}

func (t *newTranscoder) CheckFile(name string, fn func(*newSingleFileState)) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    state, found := t.files[name]
    if !found {
        return newTranscoderNotExistErr
    }
    fn(state)
    return nil
}

func (t *newTranscoder) StartShow(name, inDirPath, outParentDirPath, profile string) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    if !t.started {
        return newTranscoderNotStartedErr
    }
    state, found := t.shows[name]
    if found && (state.state == newStateInProgress || state.state == newStateNotStarted) {
        return newTranscoderAlreadyExistsErr
    }
    state = &newShowState{
        inDirPath: inDirPath,
        outParentDirPath: outParentDirPath,
        profile: profile,
        state: newStateNotStarted,
    }
    select {
    case <- t.stop:
        return newTranscoderStoppedErr
    case t.showQueue <- state:
        t.shows[name] = state
        return nil
    default:
        return newTranscoderFullErr
    }
}

func (t *newTranscoder) CheckShow(name string, fn func(*newShowState)) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    state, found := t.shows[name]
    if !found {
        return newTranscoderNotExistErr
    }
    fn(state)
    return nil
}

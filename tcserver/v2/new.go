package v2

import (
    "errors"
    "sync"

    "github.com/krelinga/video-tool-box/tcserver/hb"
)

type State int
const (
    StateUnknown State = iota
    StateNotStarted
    StateInProgress
    StateComplete
    StateError
)

type SingleFileState struct {
    // Read-only after the struct is created.
    inPath string
    outPath string
    profile string
    onDone func()  // mu will not be held when this is called.

    // Read/written concurrently.
    latest *hb.Progress
    state State
    err error

    // Must be held when reading or writing latest, state, or err.
    mu *sync.Mutex
}

// Transcodes sfs.inPath into sfs.outPath according to sfs.profile.
//
// Blocks until transcoding is finished, returning any error.
func (sfs *SingleFileState) Transcode() error {
    // TODO: implement.
    return nil
}

func TranscodeFileWorker(in <-chan *SingleFileState) {
    for work := range in {
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            work.state = StateInProgress
        }()
        err := work.Transcode()
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            if err != nil {
                work.err = err
                work.state = StateError
            } else {
                work.state = StateComplete
            }
        }()
        if work.onDone != nil {
            work.onDone()
        }
    }
}

type ShowState struct {
    // Read-only after the struct is created.
    inDirPath string
    outParentDirPath string
    profile string

    // Read/written concurrently.
    fileStates []*SingleFileState
    state State
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
func (ss *ShowState) Transcode(fileQueue chan<- *SingleFileState) error {
    // TODO: check if the output path already exists.
    // TODO: discover .mkv files & map in file paths to out file paths.
    mkvMap := make(map[string]string)
    // TODO: copy over any show-level files.
    wg := sync.WaitGroup{}
    wg.Add(len(mkvMap))
    for fromPath, toPath := range mkvMap {
        sfs := &SingleFileState{
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

func TranscodeShowWorker(in <-chan *ShowState, fileQueue chan<- *SingleFileState) {
    for work := range in {
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            work.state = StateInProgress
        }()
        err := work.Transcode(fileQueue)
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            if err != nil {
                work.err = err
                work.state = StateError
            } else {
                work.state = StateComplete
            }
        }()
    }
}

var (
    TranscoderStoppedErr = errors.New("Transcoder has been stopped")
    TranscoderAlreadyExistsErr = errors.New("Transcode already exists for this name")
    TranscoderAlreadyStartedErr = errors.New("Transcoder already started.")
    TranscoderInvalidConfigErr = errors.New("Transcoder config is invalid.")
    TranscoderNotStartedErr = errors.New("Transcoder not started.")
    TranscoderFullErr = errors.New("Transcoder is full.")
    TranscoderNotExistErr = errors.New("Transcode does not exist.")
)

type Transcoder struct {
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
    fileQueue chan *SingleFileState
    showQueue chan *ShowState

    files map[string]*SingleFileState
    shows map[string]*ShowState
    mu sync.Mutex
}

func (t *Transcoder) Start() error {
    t.mu.Lock()
    defer t.mu.Unlock()
    select {
    case <- t.stop:
        return TranscoderStoppedErr
    }
    if t.started {
        return TranscoderAlreadyStartedErr
    }
    filesCfgValid := t.FileWorkers >= 1 && t.MaxQueuedFiles >= 1
    showsCfgValid := t.ShowWorkers >= 1 && t.MaxQueuedShows >= 1
    if !(filesCfgValid && showsCfgValid) {
        return TranscoderInvalidConfigErr
    }
    t.fileQueue = make(chan *SingleFileState, t.MaxQueuedFiles)
    for i := 0; i < t.FileWorkers; i++ {
        go TranscodeFileWorker(t.fileQueue)
    }
    t.showQueue = make(chan *ShowState, t.MaxQueuedShows)
    for i := 0; i < t.ShowWorkers; i++ {
        go TranscodeShowWorker(t.showQueue, t.fileQueue)
    }
    go func() {
        <- t.stop
        close(t.fileQueue)
        close(t.showQueue)
    }()
    return nil
}

func (t *Transcoder) Stop() {
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

func (t *Transcoder) StartFile(name, inPath, outPath, profile string) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    if !t.started {
        return TranscoderNotStartedErr
    }
    state, found := t.files[name]
    if found && (state.state == StateInProgress || state.state == StateNotStarted) {
        return TranscoderAlreadyExistsErr
    }
    state = &SingleFileState{
        inPath: inPath,
        outPath: outPath,
        profile: profile,
        state: StateNotStarted,
        mu: &sync.Mutex{},
    }
    select {
    case <- t.stop:
        return TranscoderStoppedErr
    case t.fileQueue <- state:
        t.files[name] = state
        return nil
    default:
        return TranscoderFullErr
    }
}

func (t *Transcoder) CheckFile(name string, fn func(*SingleFileState)) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    state, found := t.files[name]
    if !found {
        return TranscoderNotExistErr
    }
    fn(state)
    return nil
}

func (t *Transcoder) StartShow(name, inDirPath, outParentDirPath, profile string) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    if !t.started {
        return TranscoderNotStartedErr
    }
    state, found := t.shows[name]
    if found && (state.state == StateInProgress || state.state == StateNotStarted) {
        return TranscoderAlreadyExistsErr
    }
    state = &ShowState{
        inDirPath: inDirPath,
        outParentDirPath: outParentDirPath,
        profile: profile,
        state: StateNotStarted,
    }
    select {
    case <- t.stop:
        return TranscoderStoppedErr
    case t.showQueue <- state:
        t.shows[name] = state
        return nil
    default:
        return TranscoderFullErr
    }
}

func (t *Transcoder) CheckShow(name string, fn func(*ShowState)) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    state, found := t.shows[name]
    if !found {
        return TranscoderNotExistErr
    }
    fn(state)
    return nil
}

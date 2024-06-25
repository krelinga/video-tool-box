package transcoder

import (
    "errors"
    "os"
    "path/filepath"
    "sync"

    "github.com/krelinga/video-tool-box/tcserver/hb"
    "github.com/krelinga/video-tool-box/tcserver/transcoder/related"
    "github.com/krelinga/video-tool-box/tcserver/transcoder/show"
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
    Latest *hb.Progress
    St State
    Err error

    // Must be held when reading or writing Latest, St, or Err.
    mu *sync.Mutex
}

func (sfs *SingleFileState) InPath() string {
    return sfs.inPath
}

func (sfs *SingleFileState) Profile() string {
    return sfs.profile
}

// Transcodes sfs.inPath into sfs.outPath according to sfs.profile.
//
// Blocks until transcoding is finished, returning any error.
func (sfs *SingleFileState) transcode() error {
    prog := func(u *hb.Progress) {
        sfs.mu.Lock()
        defer sfs.mu.Unlock()
        sfs.Latest = u
    }
    if err := os.MkdirAll(filepath.Dir(sfs.outPath), 0777); err != nil {
        return err
    }
    if err := related.CopyRelatedFiles(sfs.inPath, sfs.outPath); err != nil {
        return err
    }
    return hb.Run(sfs.inPath, sfs.outPath, sfs.profile, prog)
}

func transcodeFileWorker(in <-chan *SingleFileState) {
    for work := range in {
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            work.St = StateInProgress
        }()
        err := work.transcode()
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            if err != nil {
                work.Err = err
                work.St = StateError
            } else {
                work.St = StateComplete
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
    FileStates []*SingleFileState
    St State
    Err error

    mu sync.Mutex
}

func (ss *ShowState) InDirPath() string {
    return ss.inDirPath
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
func (ss *ShowState) transcode(fileQueue chan<- *SingleFileState) error {
    outDir := show.OutputDir(ss.inDirPath, ss.outParentDirPath)
    if _, err := os.Stat(outDir); errors.Is(err, os.ErrNotExist) {
        // Nothing to do here, this is the expected state.
    } else if err == nil {
        return AlreadyExistsErr
    } else {
        return err
    }

    // Find and map episodes.
    episodes, err := show.FindEpisodes(ss.inDirPath)
    if err != nil {
        return err
    }
    mkvMap := show.MapPaths(episodes, ss.inDirPath, outDir)

    // Find and map related non-episode files
    others, err := show.FindRelatedFiles(ss.inDirPath)
    if err != nil {
        return err
    }
    otherMap := show.MapPaths(others, ss.inDirPath, outDir)

    // Copy over any non-episode files.
    if err := show.CopyFiles(otherMap); err != nil {
        return err
    }

    // Transcode individual episodes.
    wg := sync.WaitGroup{}
    wg.Add(len(mkvMap))
    fileStates := make([]*SingleFileState, 0, len(mkvMap))
    for fromPath, toPath := range mkvMap {
        sfs := &SingleFileState{
            inPath: fromPath,
            outPath: toPath,
            profile: ss.profile,
            St: StateNotStarted,
            onDone: func() {
                wg.Done()
            },
            mu: &ss.mu,
        }
        fileStates = append(fileStates, sfs)
        // It's fine if we end up blocking here, although it should be rare if
        // the single-file queue is sized correctly.
        fileQueue <- sfs
    }
    func() {
        ss.mu.Lock()
        defer ss.mu.Unlock()
        ss.FileStates = fileStates
    }()

    wg.Wait()

    return nil
}

func transcodeShowWorker(in <-chan *ShowState, fileQueue chan<- *SingleFileState) {
    for work := range in {
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            work.St = StateInProgress
        }()
        err := work.transcode(fileQueue)
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            if err != nil {
                work.Err = err
                work.St = StateError
            } else {
                work.St = StateComplete
            }
        }()
    }
}

type SpreadState struct {
    // Read-only after struct is created.
    inPath string
    outParentDirPath string
    profiles []string

    // Read/written concurrently.
    FileStates []*SingleFileState
    St State
    Err error

    mu sync.Mutex
}

func (ss *SpreadState) InPath() string {
    return ss.inPath
}

// Creates one transcode for each profile in the SpreadState.
func (ss *SpreadState) transcode(fileQueue chan<- *SingleFileState) error {
    for _, p := range ss.profiles {
        outDir := filepath.Join(ss.outParentDirPath, p)
        if _, err := os.Stat(outDir); errors.Is(err, os.ErrNotExist) {
            // Nothing to do here, this is the expected state.
        } else if err == nil {
            return AlreadyExistsErr
        } else {
            return err
        }
    }

    // Create and run one transcode for each profile.
    wg := sync.WaitGroup{}
    wg.Add(len(ss.profiles))
    fileStates := make([]*SingleFileState, 0, len(ss.profiles))
    for _, p := range ss.profiles {
        outPath := filepath.Join(ss.outParentDirPath, p, filepath.Base(ss.inPath))
        sfs := &SingleFileState{
            inPath: ss.inPath,
            outPath: outPath,
            profile: p,
            St: StateNotStarted,
            onDone: func() {
                wg.Done()
            },
            mu: &ss.mu,
        }
        fileStates = append(fileStates, sfs)
        fileQueue <- sfs
    }
    func() {
        ss.mu.Lock()
        defer ss.mu.Unlock()
        ss.FileStates = fileStates
    }()

    wg.Wait()

    return nil
}

func transcodeSpreadWorker(in <-chan *SpreadState, fileQueue chan<- *SingleFileState) {
    for work := range in {
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            work.St = StateInProgress
        }()
        err := work.transcode(fileQueue)
        func() {
            work.mu.Lock()
            defer work.mu.Unlock()
            if err != nil {
                work.Err = err
                work.St = StateError
            } else {
                work.St = StateComplete
            }
        }()
    }
}

var (
    StoppedErr = errors.New("Transcoder has been stopped")
    AlreadyExistsErr = errors.New("Transcode already exists for this name")
    AlreadyStartedErr = errors.New("Transcoder already started.")
    InvalidConfigErr = errors.New("Transcoder config is invalid.")
    NotStartedErr = errors.New("Transcoder not started.")
    FullErr = errors.New("Transcoder is full.")
    NotExistErr = errors.New("Transcode does not exist.")
)

type Transcoder struct {
    // Exposed configuration variables.
    FileWorkers int
    MaxQueuedFiles int
    ShowWorkers int
    MaxQueuedShows int
    SpreadWorkers int
    MaxQueuedSpreads int

    // State management.
    started bool
    // Close when it's time to stop all processing.
    stop chan struct{}

    // processing queues.
    fileQueue chan *SingleFileState
    showQueue chan *ShowState
    spreadQueue chan *SpreadState

    files map[string]*SingleFileState
    shows map[string]*ShowState
    spreads map[string]*SpreadState
    mu sync.Mutex
}

func (t *Transcoder) Start() error {
    t.mu.Lock()
    defer t.mu.Unlock()
    if t.started {
        return AlreadyStartedErr
    }
    t.stop = make(chan struct{})
    select {
    case <- t.stop:
        return StoppedErr
    default:
        // nothing to do.
    }
    filesCfgValid := t.FileWorkers >= 1 && t.MaxQueuedFiles >= 1
    showsCfgValid := t.ShowWorkers >= 1 && t.MaxQueuedShows >= 1
    spreadsCfgValid := t.SpreadWorkers >= 1 && t.MaxQueuedSpreads >= 1
    if !(filesCfgValid && showsCfgValid && spreadsCfgValid) {
        return InvalidConfigErr
    }
    t.fileQueue = make(chan *SingleFileState, t.MaxQueuedFiles)
    for i := 0; i < t.FileWorkers; i++ {
        go transcodeFileWorker(t.fileQueue)
    }
    t.showQueue = make(chan *ShowState, t.MaxQueuedShows)
    for i := 0; i < t.ShowWorkers; i++ {
        go transcodeShowWorker(t.showQueue, t.fileQueue)
    }
    t.spreadQueue = make(chan *SpreadState, t.MaxQueuedSpreads)
    for i := 0; i < t.SpreadWorkers; i++ {
        go transcodeSpreadWorker(t.spreadQueue, t.fileQueue)
    }
    go func() {
        <- t.stop
        close(t.fileQueue)
        close(t.showQueue)
        close(t.spreadQueue)
    }()
    t.files = make(map[string]*SingleFileState)
    t.shows = make(map[string]*ShowState)
    t.spreads = make(map[string]*SpreadState)
    t.started = true
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
        return NotStartedErr
    }
    state, found := t.files[name]
    if found && (state.St == StateInProgress || state.St == StateNotStarted) {
        return AlreadyExistsErr
    }
    state = &SingleFileState{
        inPath: inPath,
        outPath: outPath,
        profile: profile,
        St: StateNotStarted,
        mu: &sync.Mutex{},
    }
    select {
    case <- t.stop:
        return StoppedErr
    case t.fileQueue <- state:
        t.files[name] = state
        return nil
    default:
        return FullErr
    }
}

func (t *Transcoder) CheckFile(name string, fn func(*SingleFileState)) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    state, found := t.files[name]
    if !found {
        return NotExistErr
    }
    state.mu.Lock()
    defer state.mu.Unlock()
    fn(state)
    return nil
}

func (t *Transcoder) StartShow(name, inDirPath, outParentDirPath, profile string) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    if !t.started {
        return NotStartedErr
    }
    state, found := t.shows[name]
    if found && (state.St == StateInProgress || state.St == StateNotStarted) {
        return AlreadyExistsErr
    }
    state = &ShowState{
        inDirPath: inDirPath,
        outParentDirPath: outParentDirPath,
        profile: profile,
        St: StateNotStarted,
    }
    select {
    case <- t.stop:
        return StoppedErr
    case t.showQueue <- state:
        t.shows[name] = state
        return nil
    default:
        return FullErr
    }
}

func (t *Transcoder) CheckShow(name string, fn func(*ShowState)) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    state, found := t.shows[name]
    if !found {
        return NotExistErr
    }
    state.mu.Lock()
    defer state.mu.Unlock()
    fn(state)
    return nil
}

func (t *Transcoder) StartSpread(name, inPath, outParentDirPath string, profiles []string) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    if !t.started {
        return NotStartedErr
    }
    state, found := t.spreads[name]
    if found && (state.St == StateInProgress || state.St == StateNotStarted) {
        return AlreadyExistsErr
    }
    state = &SpreadState{
        inPath: inPath,
        outParentDirPath: outParentDirPath,
        profiles: profiles,
        St: StateNotStarted,
    }
    select {
    case <- t.stop:
        return StoppedErr
    case t.spreadQueue <- state:
        t.spreads[name] = state
        return nil
    default:
        return FullErr
    }
}

func (t *Transcoder) CheckSpread(name string, fn func(*SpreadState)) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    state, found := t.spreads[name]
    if !found {
        return NotExistErr
    }
    state.mu.Lock()
    defer state.mu.Unlock()
    fn(state)
    return nil
}

type Type int
const (
    TypeUnknown Type = iota
    TypeSingleFile
    TypeShow
    TypeSpread
)

type Operation struct {
    Name string
    Typ Type
    St State
}

func (t *Transcoder) List() []*Operation {
    t.mu.Lock()
    defer t.mu.Unlock()
    out := make([]*Operation, 0, len(t.files) + len(t.shows) + len(t.spreads))
    for name, op := range t.files {
        op.mu.Lock()
        defer op.mu.Unlock()
        out = append(out, &Operation{
            Name: name,
            Typ: TypeSingleFile,
            St: op.St,
        })
    }
    for name, op := range t.shows {
        op.mu.Lock()
        defer op.mu.Unlock()
        out = append(out, &Operation{
            Name: name,
            Typ: TypeShow,
            St: op.St,
        })
    }
    for name, op := range t.spreads {
        op.mu.Lock()
        defer op.mu.Unlock()
        out = append(out, &Operation{
            Name: name,
            Typ: TypeSpread,
            St: op.St,
        })
    }
    return out
}

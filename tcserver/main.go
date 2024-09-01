package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/krelinga/video-tool-box/tcserver/transcoder"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	pbconnect "buf.build/gen/go/krelinga/proto/connectrpc/go/krelinga/video/tcserver/v1/tcserverv1connect"
)

func getPort() (int, error) {
	const envVar = "VTB_TCSERVER_PORT"
	portString, err := getEnvVar(envVar)
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return 0, fmt.Errorf("env var %s should be a port number, saw %s", envVar, portString)
	}
	return port, nil
}

func mainOrError() error {
	fmt.Println("hello world!")

	tran := transcoder.Transcoder{}
	var err error
	tran.FileWorkers, err = getEnvVarInt("VTB_TCSERVER_FILE_WORKERS")
	if err != nil {
		return err
	}
	tran.MaxQueuedFiles, err = getEnvVarInt("VTB_TCSERVER_MAX_QUEUED_FILES")
	if err != nil {
		return err
	}
	tran.ShowWorkers, err = getEnvVarInt("VTB_TCSERVER_SHOW_WORKERS")
	if err != nil {
		return err
	}
	tran.MaxQueuedShows, err = getEnvVarInt("VTB_TCSERVER_MAX_QUEUED_SHOWS")
	if err != nil {
		return err
	}
	tran.SpreadWorkers, err = getEnvVarInt("VTB_TCSERVER_SPREAD_WORKERS")
	if err != nil {
		return err
	}
	tran.MaxQueuedSpreads, err = getEnvVarInt("VTB_TCSERVER_MAX_QUEUED_SPREADS")
	if err != nil {
		return err
	}

	if err = tran.Start(); err != nil {
		return err
	}
	defer tran.Stop()

	profile, err := getEnvVar("VTB_TCSERVER_PROFILE")
	if err != nil {
		return err
	}
	port, err := getPort()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	path, handler := pbconnect.NewTCServiceHandler(newTcServer(profile, &tran))
	mux.Handle(path, handler)
	// Runs as long as the server is alive.
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), h2c.NewHandler(mux, &http2.Server{}))

	return nil
}

func main() {
	if err := mainOrError(); err != nil {
		log.Fatal(err)
	}
}

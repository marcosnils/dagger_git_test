package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	daemon "github.com/aymanbagabas/go-git-daemon"
)

type gitServer struct {
	server *daemon.Server
}

func NewGitServer() (*gitServer, error) {
	logger := log.Default()

	logger.SetPrefix("[git-daemon]")
	server := &daemon.Server{
		BasePath:             ".",
		Logger:               logger,
		ExportAll:            true,
		Verbose:              true,
		UploadPackHandler:    daemon.DefaultUploadPackHandler,
		UploadArchiveHandler: daemon.DefaultUploadArchiveHandler,
		ReceivePackHandler:   daemon.DefaultReceivePackHandler,
	}

	return &gitServer{server: server}, nil
}

func (gs *gitServer) Start() (serr error) {
	unixListener, err := net.Listen("unix", "git-server.sock")
	if err != nil {
		serr = err
		return
	}

	go func() {
		signalChan := make(chan os.Signal, 1)

		signal.Notify(
			signalChan,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGQUIT,
		)
		<-signalChan
		log.Print("os.Interrupt - shutting down...\n")

		gs.Stop()
	}()

	var done chan struct{}
	// start serving
	go func() {
		fmt.Println("Starting server in unix://git-server.sock")
		// Start HTTP server
		if err := gs.server.Serve(unixListener); err != nil {
			// if err := gs.server.ListenAndServe(":12345"); err != nil {
			done <- struct{}{}
			serr = err
		}
	}()

	<-done

	return
}

func (gs *gitServer) Stop() {
	if err := gs.server.Close(); err != nil {
		log.Printf("shutdown error: %v\n", err)
		defer os.Exit(1)
	} else {
		log.Printf("gracefully stopped\n")
	}
}

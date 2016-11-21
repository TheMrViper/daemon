package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/TheMrViper/daemon"
)

type TestService struct {
	toClose chan bool
}

func NewTestService() *TestService {
	return &TestService{
		toClose: make(chan bool),
	}
}

func (s *TestService) Start() error {
	for {
		select {
		case <-s.toClose:
			break
		case <-time.Tick(10 * time.Second):
			fmt.Println(time.Now())
		}
	}
	return nil
}

func (s *TestService) Shutdown() error {
	s.toClose <- true
	return nil
}

func (s *TestService) Terminate() error {
	return nil
}

var (
	stopFLag    = flag.Bool("stop", false, "")
	startFlag   = flag.Bool("start", false, "")
	restartFlag = flag.Bool("restart", false, "")
)

func main() {
	config := daemon.Config{}

	service := daemon.New(config)
	if *startFlag {

		service.Func(syscall.SIGABRT, customHandler)
		service.Command(daemon.BoolFlag(stopFLag, true), syscall.SIGQUIT)
		service.Command(daemon.BoolFlag(restartFlag, true), syscall.SIGHUP)

		service.RestartSignals(syscall.SIGHUP)
		service.ShutdownSignals(syscall.SIGQUIT)
		service.TerminateSignals(syscall.SIGTERM)

		service.AddService(NewTestService())

		if err := service.Start(); err != nil {
			fmt.Println("Daemon error: ", err)
			return
		}
		fmt.Println("Daemon started")
	}
}

func customHandler(sig os.Signal) error {
	fmt.Println("hello custom handler")
	return nil
}

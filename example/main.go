package main

import (
	"time"
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
		case <-toCLose:
			break
		case <-time.Tick(10 * time.Second):
			fmt.Println(time.Now())
		}
	}
	return nil
}

func (s *TestService) Graceful() error {
	s.toClose <- true
	return nil
}

func (s *TestService) Terminate() error {
	return nil
}

var (
	stop = flag.BoolFlag("stop", false, "")
	start = flag.BoolFlag("start", false, "")
	restart = flag.BoolFlag("restart", false, "")
)

func main() {
	config := daemon.Config{}
	
	
	daemon := daemon.New(config)
	if *start {
		daemon.Func(syscall.SIGTEST, sample)
		daemon.Command(daemon.BoolFlag(flag, true), syscall.SIGHUP)
	
		daemon.RestartSignals(...)
		daemon.ShutdownSignals(...)
		daemon.TerminateSignals(...)
	
		daemon.AddService(NewTestService())
		
		if err := daemon.Start();err!=nil {
			fmt.Println("Daemon error: ", err)
			return
		}
		fmt.Println("Daemon started")
	}
}
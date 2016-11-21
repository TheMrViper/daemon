package daemon

import (
	"os"

	"github.com/TheMrViper/daemon/pid"
)

type (
	Daemon struct {
		config Config

		abspath  string
		pidFile  *pid.File
		logFile  *os.File
		nullFile *os.File

		rpipe, wpipe *os.File

		services []Service

		cmds  map[Flag]os.Signal
		funcs map[os.Signal]HandlerFunc

		restartSignals   []os.Signal
		shutdownSignals  []os.Signal
		terminateSignals []os.Signal
	}
	Service interface {
		Start() error
		Shutdown() error
		Terminate() error
	}
	HandlerFunc func(os.Signal) error
)

func IsChild() bool {
	return os.Getenv(MARK_NAME) == MARK_VALUE
}

func (d *Daemon) Start() (err error) {
	if !IsChild() {
		if d.toSend() {
			child, err := d.search()
			if err != nil {
				return err
			}
			if child != nil {
				return d.sendSignals(child)
			}
		}

		return d.fork()
	}

	if err := d.work(); err != nil {
		return err
	}
	defer d.kill()

	return d.recvSignals()
}

func (d *Daemon) AddService(s Service) {
	d.services = append(d.services, s)
}

func (d *Daemon) Func(sig os.Signal, h HandlerFunc) {
	d.funcs[sig] = h
}
func (d *Daemon) Command(flag Flag, sig os.Signal) {
	d.cmds[flag] = sig
}

func (d *Daemon) RestartSignals(sigs ...os.Signal) {
	d.restartSignals = append(d.restartSignals, sigs...)
}
func (d *Daemon) ShutdownSignals(sigs ...os.Signal) {
	d.shutdownSignals = append(d.shutdownSignals, sigs...)
}
func (d *Daemon) TerminateSignals(sigs ...os.Signal) {
	d.terminateSignals = append(d.terminateSignals, sigs...)
}
